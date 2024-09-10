package moqt

import (
	"fmt"
	"math/rand"
	"moq-go/moqt/wire"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

var DEFAULT_SERVER_SETUP = wire.ServerSetup{SelectedVersion: wire.DRAFT_04, Params: wire.Parameters{
	wire.ROLE_PARAM: &wire.IntParameter{Ptype: wire.ROLE_PARAM, Pvalue: wire.ROLE_RELAY},
}}

var sm *SessionManager = NewSessionManager()

type MOQTConnection interface {
	AcceptStream(context context.Context) (quic.Stream, error)
	AcceptUniStream(context context.Context) (quic.ReceiveStream, error)
	CloseWithError(quic.ApplicationErrorCode, string) error
	OpenUniStreamSync(ctx context.Context) (quic.SendStream, error)
	OpenUniStream() (quic.SendStream, error)
}

type MOQTSession struct {
	Conn              MOQTConnection
	CS                *ControlStream
	ctx               context.Context
	id                string
	RemoteRole        uint64
	LocalRole         uint64
	cancelFunc        func()
	Slogger           zerolog.Logger
	Handler           Handler
	IncomingStreams   StreamsMap
	SubscribedStreams StreamsMap
}

func CreateMOQSession(conn MOQTConnection, LocalRole uint64) (*MOQTSession, error) {
	session := &MOQTSession{}
	session.Conn = conn
	session.ctx, session.cancelFunc = context.WithCancel(context.Background())
	session.id = strings.Split(uuid.New().String(), "-")[0]
	session.RemoteRole = 0
	session.LocalRole = LocalRole
	session.IncomingStreams = NewStreamsMap(session)
	session.SubscribedStreams = NewStreamsMap(session)

	session.Slogger = log.With().Str("ID", session.id).Str("Role", wire.GetRoleStringVarInt(session.RemoteRole)).Logger()

	if handler, err := CreateNewHandler(LocalRole, session); err != nil {
		return nil, err
	} else {
		session.Handler = handler
	}

	sm.addSession(session)

	return session, nil
}

func (s *MOQTSession) Close(code uint64, msg string) {
	s.Conn.CloseWithError(quic.ApplicationErrorCode(code), msg)
	s.cancelFunc()

	s.SubscribedStreams.DeleteAll()

	s.Slogger.Error().Msgf("[%s][Closing MOQT Session][Code - %d]%s", s.id, code, msg)

	sm.removeSession(s)
}

func (s *MOQTSession) Serve() {
	go s.handleControlStream()
}

func (s *MOQTSession) SetRemoteRole(role uint64) {
	s.RemoteRole = role
	s.Slogger = log.With().Str("ID", s.id).Str("Role", wire.GetRoleStringVarInt(s.RemoteRole)).Logger()

	if s.RemoteRole == wire.ROLE_PUBLISHER || s.RemoteRole == wire.ROLE_RELAY {
		go s.handleUniStreams()
	}
}

// Fetches the Object Stream with the StreamID (OR) Forwards the Subscribe and returns the ObjectStream Placeholder
func (s *MOQTSession) GetObjectStream(msg *wire.SubscribeMessage) *ObjectStream {

	streamid := msg.GetStreamID()
	stream, found := s.IncomingStreams.StreamIDGetStream(streamid)

	// We need to fetch the fresh copies of ".catalog", "audio.mp4", "video.mp4".I knowm tts a nasty implementation. Requires more work.
	if !found || msg.ObjectStreamName == ".catalog" || msg.ObjectStreamName == "audio.mp4" || msg.ObjectStreamName == "video.mp4" {
		subid := uint64(rand.Uint32())
		stream = s.IncomingStreams.CreateNewStream(subid, streamid)

		submsg := *msg
		submsg.SubscribeID = subid
		s.CS.WriteControlMessage(&submsg)
	}

	return stream
}

func (s *MOQTSession) DispatchObject(object *MOQTObject) {
	unistream, err := s.Conn.OpenUniStream()

	if err != nil {
		s.Slogger.Error().Msgf("[Error Opening Unistream][%s]", err)
		return
	}

	streamid := object.header.GetTrackAlias()
	subid := s.SubscribedStreams.subidmap[streamid]

	groupHeader := object.header
	unistream.Write(groupHeader.GetBytes(subid))

	reader := object.NewReader()
	reader.Pipe(unistream)

	unistream.Close()
}

// Stream Handlers
func (s *MOQTSession) handleControlStream() {

	for {
		var err error
		stream, err := s.Conn.AcceptStream(s.ctx)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			s.Close(wire.MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Session Closed]"))
			return
		}

		if err != nil {
			s.Close(wire.MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Error Accepting Control Stream][%s]", err))
			return
		}

		if s.CS != nil {
			s.Close(wire.MOQERR_PROTOCOL_VIOLATION, "Received Control Stream Twice")
			return
		}

		s.CS = NewControlStream(s, stream)
		go s.CS.Serve()
	}
}

func (s *MOQTSession) handleUniStreams() {

	for {
		unistream, err := s.Conn.AcceptUniStream(s.ctx)

		if err != nil {
			log.Error().Msgf("[Error Acceping Unistream][%s]", err)
		}

		go func(stream quic.ReceiveStream) {

			reader := quicvarint.NewReader(stream)
			header, err := wire.ParseMOQTObjectHeader(reader)

			if err != nil {
				s.Slogger.Error().Msgf("[Error Parsing Object Header][%s]", err)
				return
			}

			object := NewMOQTObject(header)
			go object.ParseFromStream(reader)

			subid := header.GetSubID()

			if objectStream, found := s.IncomingStreams.SubIDGetStream(subid); found {
				objectStream.AddObject(object)
			} else {
				s.Slogger.Error().Msgf("[Object Stream Not Found][Alias - %d]", object.header.GetTrackAlias())
			}

		}(unistream)
	}
}

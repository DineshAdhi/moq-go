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

var DEFAULT_CLIENT_SETUP = wire.ClientSetup{
	SupportedVersions: []uint64{wire.DRAFT_04},
	Params: wire.Parameters{
		wire.ROLE_PARAM: &wire.IntParameter{Ptype: wire.ROLE_PARAM, Pvalue: wire.ROLE_RELAY},
	},
}

var sm *SessionManager = NewSessionManager()

type MOQTConnection interface {
	AcceptStream(context context.Context) (quic.Stream, error)
	AcceptUniStream(context context.Context) (quic.ReceiveStream, error)
	CloseWithError(quic.ApplicationErrorCode, string) error
	OpenUniStreamSync(ctx context.Context) (quic.SendStream, error)
	OpenUniStream() (quic.SendStream, error)
	OpenStream() (quic.Stream, error)
}

const (
	SERVER_MODE = uint8(0x34)
	CLIENT_MODE = uint8(0x66)
)

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
	Mode              uint8
}

func CreateMOQSession(conn MOQTConnection, LocalRole uint64, mode uint8) (*MOQTSession, error) {
	session := &MOQTSession{}
	session.Conn = conn
	session.ctx, session.cancelFunc = context.WithCancel(context.Background())
	session.id = strings.Split(uuid.New().String(), "-")[0]
	session.RemoteRole = 0
	session.LocalRole = LocalRole
	session.IncomingStreams = NewStreamsMap(session)
	session.SubscribedStreams = NewStreamsMap(session)
	session.Mode = mode

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

func (s *MOQTSession) SendSubscribeOk(streamid string, okm *wire.SubscribeOk) {
	if subid, ok := s.SubscribedStreams.streamidmap[streamid]; ok {
		subokmsg := *okm
		subokmsg.SubscribeID = subid

		s.CS.WriteControlMessage(&subokmsg)
	}
}

func (s *MOQTSession) SendUnsubscribe(subid uint64) {
	msg := &wire.Unsubcribe{
		SubscriptionID: subid,
	}

	s.CS.WriteControlMessage(msg)
}

func (s *MOQTSession) ServeMOQ() {
	switch s.Mode {
	case SERVER_MODE:
		go s.handleControlStream()
	case CLIENT_MODE:
		go s.InitiateHandshake()
	}
}

func (s *MOQTSession) SetRemoteRole(role uint64) {
	s.RemoteRole = role
	s.Slogger = log.With().Str("ID", s.id).Str("RemoteRole", wire.GetRoleStringVarInt(s.RemoteRole)).Logger()

	switch s.LocalRole {
	case wire.ROLE_RELAY:
	}

	if s.RemoteRole == wire.ROLE_PUBLISHER || s.RemoteRole == wire.ROLE_RELAY {
		go s.handleUniStreams()
	}
}

// Fetches the Object Stream with the StreamID (OR) Forwards the Subscribe and returns the ObjectStream Placeholder
func (s *MOQTSession) GetObjectStream(msg *wire.Subscribe) *ObjectStream {

	streamid := msg.GetStreamID()
	stream, found := s.IncomingStreams.StreamIDGetStream(streamid)

	// We need to fetch the fresh copies of ".catalog", "audio.mp4", "video.mp4".I knowm its a nasty implementation. Requires more work.
	if !found || strings.Contains(msg.TrackName, ".catalog") || strings.Contains(msg.TrackName, ".mp4") {
		subid := uint64(rand.Uint32())
		stream = s.IncomingStreams.CreateNewStream(subid, streamid)

		submsg := *msg
		submsg.SubscribeID = subid
		s.CS.WriteControlMessage(&submsg)
	}

	return stream
}

func (s *MOQTSession) isUpstream() bool {
	return s.RemoteRole == wire.ROLE_PUBLISHER || s.RemoteRole == wire.ROLE_RELAY
}

func (s *MOQTSession) DispatchObject(object *MOQTObject) {

	if subid, ok := s.SubscribedStreams.streamidmap[object.GetStreamID()]; ok {

		unistream, err := s.Conn.OpenUniStream()

		if err != nil {
			s.Slogger.Error().Msgf("[Error Opening Unistream][%s]", err)
			return
		}

		groupHeader := object.header
		unistream.Write(groupHeader.GetBytes(subid))

		reader := object.NewReader()
		reader.Pipe(unistream)

		unistream.Close()
	} else {
		s.Slogger.Error().Msgf("[Unable to find DownStream SubID for StreamID][Stream ID - %s]", object.GetStreamID())
	}
}

func (s *MOQTSession) InitiateHandshake() {
	cs, err := s.Conn.OpenStream()

	if err != nil {
		s.Close(wire.MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Handshake Failed][Error Opening ControlStream]"))
		return
	}

	s.CS = NewControlStream(s, cs)
	go s.CS.ServeCS()

	s.CS.WriteControlMessage(&DEFAULT_CLIENT_SETUP)
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
		go s.CS.ServeCS()
	}
}

func (s *MOQTSession) handleUniStreams() {

	for {
		unistream, err := s.Conn.AcceptUniStream(s.ctx)

		if err != nil {
			log.Error().Msgf("[Error Acceping Unistream][%s]", err)
			break
		}

		go func(stream quic.ReceiveStream) {

			reader := quicvarint.NewReader(stream)
			header, err := wire.ParseMOQTObjectHeader(reader)

			if err != nil {
				s.Slogger.Error().Msgf("[Error Parsing Object Header][%s]", err)
				return
			}

			subid := header.GetSubID()

			if objectStream, found := s.IncomingStreams.SubIDGetStream(subid); found {

				object := NewMOQTObject(header, objectStream.streamid)
				go object.ParseFromStream(reader)

				objectStream.AddObject(object)

			} else {
				s.Slogger.Error().Msgf("[Object Stream Not Found][Alias - %d]", header.GetTrackAlias())
			}

		}(unistream)
	}
}

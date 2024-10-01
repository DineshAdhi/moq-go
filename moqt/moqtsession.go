package moqt

import (
	"fmt"
	"strings"

	"github.com/DineshAdhi/moq-go/moqt/wire"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

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
	Conn          MOQTConnection
	CS            *ControlStream
	ctx           context.Context
	Id            string
	RemoteRole    uint64
	LocalRole     uint64
	cancelFunc    func()
	Slogger       zerolog.Logger
	Handler       Handler
	Mode          uint8
	HandshakeDone chan bool
}

func CreateMOQSession(conn MOQTConnection, LocalRole uint64, mode uint8) (*MOQTSession, error) {

	session := &MOQTSession{}
	session.Conn = conn
	session.ctx, session.cancelFunc = context.WithCancel(context.Background())
	session.Id = strings.Split(uuid.New().String(), "-")[0]
	session.RemoteRole = 0
	session.LocalRole = LocalRole
	session.Mode = mode
	session.HandshakeDone = make(chan bool, 1)

	session.Slogger = log.With().Str("ID", session.Id).Str("Role", wire.GetRoleStringVarInt(session.RemoteRole)).Logger()

	if handler, err := CreateNewHandler(LocalRole, session); err != nil {
		return nil, err
	} else {
		session.Handler = handler
	}

	sm.addSession(session)

	return session, nil
}

func (s *MOQTSession) isUpstream() bool {
	return s.RemoteRole == wire.ROLE_PUBLISHER || s.RemoteRole == wire.ROLE_RELAY
}

func (s *MOQTSession) Close(code uint64, msg string) {
	s.Conn.CloseWithError(quic.ApplicationErrorCode(code), msg)
	s.cancelFunc()

	s.Slogger.Error().Msgf("[%s][Closing MOQT Session][Code - %d]%s", s.Id, code, msg)

	s.Handler.HandleClose()

	sm.removeSession(s)
}

func (s *MOQTSession) ServeMOQ() {
	switch s.Mode {
	case SERVER_MODE:
		go s.handleControlStream()
	case CLIENT_MODE:
		go s.InitiateHandshake()
	}

	go s.Handler.DoHandle() // ObjectStream Processing happens on the respective handlers (relay / pub / sub).
}

func (s *MOQTSession) SendSubscribe(submsg *wire.Subscribe) {
	s.CS.WriteControlMessage(submsg)
}

func (s *MOQTSession) SendUnsubscribe(subid uint64) {
	msg := &wire.Unsubcribe{
		SubscriptionID: subid,
	}

	s.CS.WriteControlMessage(msg)
}

func (s *MOQTSession) InitiateHandshake() {
	cs, err := s.Conn.OpenStream()

	if err != nil {
		s.Close(wire.MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Handshake Failed][Error Opening ControlStream]"))
		return
	}

	s.CS = NewControlStream(s, cs)
	go s.CS.ServeCS()

	clientSetup := wire.ClientSetup{
		SupportedVersions: []uint64{wire.DRAFT_04},
		Params: wire.Parameters{
			wire.ROLE_PARAM: &wire.IntParameter{Ptype: wire.ROLE_PARAM, Pvalue: s.LocalRole},
		},
	}

	s.CS.WriteControlMessage(&clientSetup)
}

// This function sets the remote role of the session and starts the listeners accordingly. This is crucial part of the MOQT Handshake.
func (s *MOQTSession) SetRemoteRole(role uint64) error {
	s.RemoteRole = role
	s.Slogger = log.With().Str("ID", s.Id).Str("RemoteRole", wire.GetRoleStringVarInt(s.RemoteRole)).Logger()

	switch s.LocalRole {
	case wire.ROLE_PUBLISHER:
		if s.RemoteRole == wire.ROLE_PUBLISHER {
			return fmt.Errorf("I am a Pub. Role Error")
		}
	case wire.ROLE_SUBSCRIBER:
		if s.RemoteRole == wire.ROLE_SUBSCRIBER {
			return fmt.Errorf("I am a Sub. Role Error")
		}
	case wire.ROLE_RELAY:
	default:
		return fmt.Errorf("Unknown Role")
	}

	return nil
}

func (s *MOQTSession) RelayHandler() *RelayHandler {
	return s.Handler.(*RelayHandler)
}

func (s *MOQTSession) PubHandler() *PubHandler {
	return s.Handler.(*PubHandler)
}

func (s *MOQTSession) SubHandler() *SubHandler {
	return s.Handler.(*SubHandler)
}

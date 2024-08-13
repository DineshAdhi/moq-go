package moqt

import (
	"fmt"
	"moq-go/logger"
	"strings"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	"golang.org/x/net/context"
)

var DEFAULT_SERVER_SETUP = ServerSetup{SelectedVersion: DRAFT_03, Params: Parameters{
	ROLE_PARAM: &IntParameter{ROLE_PARAM, ROLE_PUBSUB},
}}

var sm SessionManager = NewSessionManager()

type MOQTConnection interface {
	AcceptStream(context context.Context) (quic.Stream, error)
	AcceptUniStream(context context.Context) (quic.ReceiveStream, error)
	CloseWithError(quic.ApplicationErrorCode, string) error
}

type MOQTSession struct {
	Conn            MOQTConnection
	controlStream   quic.Stream
	ctx             context.Context
	ishandshakedone bool
	id              string
	role            uint64
	cancelFunc      func()
}

func CreateMOQSession(conn MOQTConnection) *MOQTSession {
	session := &MOQTSession{}
	session.Conn = conn
	session.ctx, session.cancelFunc = context.WithCancel(context.Background())
	session.id = strings.Split(uuid.New().String(), "-")[0]

	sm.addSession(session)

	return session
}

func (s *MOQTSession) Close(code uint64, msg string) {
	s.Conn.CloseWithError(quic.ApplicationErrorCode(code), msg)
	s.cancelFunc()

	sm.removeSession(s)

	logger.ErrorLog("[Closing MOQT Session][Code - %d]%s", code, msg)
}

func (s *MOQTSession) WriteControlMessage(msg MOQTMessage) {
	logger.DebugLog("[%s][Dipsatching CONTROL]%s", s.id, msg.String())
	s.controlStream.Write(msg.GetBytes())
}

func (s MOQTSession) Serve() {

	var err error
	s.controlStream, err = s.Conn.AcceptStream(s.ctx)

	if err != nil {
		s.Close(MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Error Accepting Control Stream][%s]", err))
	}

	controlHandler := NewControlHandler(&s)
	go controlHandler.Run()

	objStream, err := s.Conn.AcceptUniStream(s.ctx)

	if err != nil {
		s.Close(MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Error Accepting Object Stream][%s]", err))
		return
	}

	logger.DebugLog("[OBJECT Stream Acquired][%+v]", objStream)
}

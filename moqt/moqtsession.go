package moqt

import (
	"moq-go/logger"
	"sync"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"golang.org/x/net/context"
)

var DEFAULT_SERVER_SETUP = ServerSetup{SelectedVersion: DRAFT_03, Params: Parameters{
	ROLE_PARAM: &IntParameter{ROLE_PARAM, ROLE_PUBSUB},
}}

var sm SessionManager = SessionManager{rwlock: sync.RWMutex{}, sessions: map[string]*MOQTSession{}}

type MOQTConnection interface {
	AcceptStream(context context.Context) (quic.Stream, error)
	AcceptUniStream(context context.Context) (quic.ReceiveStream, error)
	CloseWithError(quic.ApplicationErrorCode, string) error
}

type MOQTSession struct {
	id            string
	Conn          MOQTConnection
	ControlStream quic.Stream
	handshakedone bool
	namespaces    map[string]map[uint64]string
	sentSubscribe bool
	role          uint64
	ctx           context.Context
	cancelFunc    func()
}

func CreateMOQSession(conn MOQTConnection) *MOQTSession {
	session := &MOQTSession{}
	session.Conn = conn
	session.namespaces = make(map[string]map[uint64]string)
	session.sentSubscribe = false
	session.ctx, session.cancelFunc = context.WithCancel(context.Background())
	session.id = uuid.New().String()

	sm.addSession(session)

	return session
}

func (s *MOQTSession) Serve() {
	go s.handleControlStream()
}

func (s MOQTSession) Close(code uint64, msg string) {
	s.Conn.CloseWithError(quic.ApplicationErrorCode(code), msg)
}

func (s *MOQTSession) handleControlStream() {

	var err error
	s.ControlStream, err = s.Conn.AcceptStream(s.ctx)

	if err != nil {
		logger.ErrorLog("[Error Accepting Control Stream from WTS][%s]", err)
		s.Close(MOQERR_INTERNAL_ERROR, "Unable to accept Control Stream")
		sm.removeSession(s)
		return
	}

	controlReader := quicvarint.NewReader(s.ControlStream)

	for {
		msg, err := ParseMOQTMessage(controlReader)

		logger.DebugLog("[%s][MOQT Message Parsed]%+v", s.id, msg)

		if err != nil {
			logger.ErrorLog("[Error Parsing Control Message][%s]", err)
			s.Conn.CloseWithError(quic.ApplicationErrorCode(MOQERR_INTERNAL_ERROR), "Internal Error")
			sm.removeSession(s)
			return
		}

		if s.handshakedone {
			s.handleControlMessage(msg)
		} else {
			s.handleSetupMessage(msg)
		}
	}
}

func (s *MOQTSession) handleSetupMessage(msg MOQTMessage) {

	switch msg.Type() {
	case CLIENT_SETUP:
		s.handleClientSetup(msg)
	default:
	}

}

func (s *MOQTSession) handleClientSetup(msg MOQTMessage) {
	clientSetup := msg.(*ClientSetup)

	if !clientSetup.CheckDraftSupport() {
		logger.ErrorLog("[MOQT Handshake Error][Draft Version not Supported][%+v]", clientSetup.SupportedVersions)
		s.Conn.CloseWithError(quic.ApplicationErrorCode(MOQERR_INTERNAL_ERROR), "Internal Error")
		return
	}

	for key, param := range clientSetup.Params {
		if key == ROLE_PARAM {
			s.role = param.Value().(uint64)
			s.id = s.id + "_" + GetRoleString(param)
		}
	}

	serverSetup := DEFAULT_SERVER_SETUP
	s.ControlStream.Write(serverSetup.GetBytes())

	s.handshakedone = true
}

func (s *MOQTSession) handleControlMessage(msg MOQTMessage) {

	switch msg.Type() {
	case ANNOUNCE:
		s.handleAnnounce(msg)
	case SUBSCRIBE:
		s.handleSubscribe(msg)
	case SUBSCRIBE_OK:
		s.handleSubscribeOk(msg)
	default:
	}
}

func (s *MOQTSession) handleAnnounce(msg MOQTMessage) {
	announceMsg := msg.(*AnnounceMessage)
	announceOk := AnnounceOkMessage{}
	announceOk.GetAnnounceOk(announceMsg)

	s.namespaces[announceMsg.tracknamespace] = map[uint64]string{}

	s.ControlStream.Write(announceOk.GetBytes())
}

func (s *MOQTSession) handleSubscribe(msg MOQTMessage) {
	subcribeMessage := msg.(*SubscribeMessage)
	ps, err := sm.getSessionWithNamespace(subcribeMessage.TrackNamespace)

	if err != nil {
		logger.ErrorLog("[Error Processing Subscribe][%s]", err)
		return
	}

	logger.DebugLog("[Forwarding Subscribe][%s --> %s]", s.id, ps.id)

	s.sentSubscribe = true
	ps.forwardSubscribe(subcribeMessage)
}

func (s *MOQTSession) handleSubscribeOk(msg MOQTMessage) {
	subokmsg := msg.(*SubsribeOkMessage)
	sm.forwardSubscribeOk(subokmsg)
}

func (s *MOQTSession) forwardSubscribe(subcribeMessage *SubscribeMessage) {
	s.ControlStream.Write(subcribeMessage.GetBytes())
}

func (s *MOQTSession) forwardSubscribeOk(subok *SubsribeOkMessage) {

	if s.sentSubscribe {
		s.ControlStream.Write(subok.GetBytes())
		logger.DebugLog("[%s][Forwarding SubscribeOk]", s.id)
		s.sentSubscribe = false
	}
}

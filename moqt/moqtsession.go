package moqt

import (
	"moq-go/logger"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"golang.org/x/net/context"
)

var DEFAULT_SERVER_SETUP = ServerSetup{SelectedVersion: DRAFT_03, Params: Parameters{
	ROLE_PARAM: &IntParameter{ROLE_PARAM, ROLE_PUBSUB},
}}

// TODO : Write MOQT session generically to support both WTransPort & QUIC

type MOQTConnection interface {
	AcceptStream(context context.Context) (quic.Stream, error)
	AcceptUniStream(context context.Context) (quic.ReceiveStream, error)
	CloseWithError(quic.ApplicationErrorCode, string) error
}

type MOQTSession struct {
	Conn          MOQTConnection
	controlStream quic.Stream
	handshakedone bool
}

func (s MOQTSession) Serve() {

	controlStream, err := s.Conn.AcceptStream(context.TODO())

	if err != nil {
		logger.ErrorLog("[Error Accepting Control Stream from WTS][%s]", err)
		s.Conn.CloseWithError(quic.ApplicationErrorCode(MOQERR_INTERNAL_ERROR), "Internal Error")
		return
	}

	s.controlStream = controlStream

	go s.handleControlStream()
}

func (s *MOQTSession) handleControlStream() {

	controlReader := quicvarint.NewReader(s.controlStream)

	for {
		msg, err := ParseMOQTMessage(controlReader)

		if err != nil {
			logger.ErrorLog("[Error Parsing Control Message][%s]", err)
			s.Conn.CloseWithError(quic.ApplicationErrorCode(MOQERR_INTERNAL_ERROR), "Internal Error")
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

	serverSetup := DEFAULT_SERVER_SETUP
	s.controlStream.Write(serverSetup.GetBytes())

	s.handshakedone = true
}

func (s *MOQTSession) handleControlMessage(msg MOQTMessage) {

	switch msg.Type() {
	case ANNOUNCE:
		logger.DebugLog("[Got Announce][%+v]", msg)
	case SUBSCRIBE:
		logger.DebugLog("[Got Subscribe][%+v]", msg)
	default:
	}
}

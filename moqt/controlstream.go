package moqt

import (
	"fmt"
	"moq-go/moqt/wire"
	"net"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

type ControlStream struct {
	*MOQTSession
	stream          quic.Stream
	ishandshakedone bool
}

func NewControlStream(session *MOQTSession, stream quic.Stream) *ControlStream {
	cs := &ControlStream{session, stream, false}
	return cs
}

func (cs *ControlStream) WriteControlMessage(msg wire.MOQTMessage) {

	if cs.stream == nil {
		cs.Slogger.Error().Msgf("[Error Writing Control Message][CS is nil][HS - %+v]", cs.ishandshakedone)
		return
	}

	_, err := cs.stream.Write(msg.GetBytes())

	if err != nil {
		cs.Slogger.Error().Msgf("[Error Writing to Control][%s]", err)
	}

	cs.Slogger.Debug().Msgf("[Dispatching CONTROL]%s", msg.String())
}

func (cs *ControlStream) ServeCS() {

	reader := quicvarint.NewReader(cs.stream)

	for {
		moqtMessage, err := wire.ParseMOQTMessage(reader)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			return
		}

		if err != nil {
			log.Debug().Msgf("[Error Parsing MOQT Message][%s]", err)
			break
		}

		if cs.ishandshakedone {
			cs.handleControlMessage(moqtMessage)
		} else {
			cs.handleSetupMessage(moqtMessage)
		}
	}
}

func (cs *ControlStream) handleSetupMessage(m wire.MOQTMessage) {

	var role uint64
	var err error

	switch m.Type() {
	case wire.CLIENT_SETUP:

		clientSetup := m.(*wire.ClientSetup)

		cs.Slogger.Debug().Msgf(clientSetup.String())

		if !clientSetup.CheckDraftSupport() {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "CLIENT SETUP ERROR : PROTOCOL DRAFT NOT SUPPORTED")
			return
		}

		if role, err = clientSetup.GetRoleParam(); err != nil {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "CLIENT SETUP ERROR : ROLE PARAM MISSING")
			return
		}

		if err := cs.SetRemoteRole(role); err != nil {
			cs.Close(wire.MOQERR_PROTOCOL_VIOLATION, err.Error())
			return
		}

		cs.ishandshakedone = true
		cs.HandshakeDone <- true

		serverSetup := wire.NewServerSetup(wire.DRAFT_04, wire.Parameters{
			wire.ROLE_PARAM: wire.NewIntParameter(wire.ROLE_PARAM, cs.LocalRole),
		})

		cs.WriteControlMessage(&serverSetup)

		cs.Slogger.Info().Msgf("[Handshake Success]")

	case wire.SERVER_SETUP:

		serverSetup := m.(*wire.ServerSetup)

		cs.Slogger.Debug().Msgf(serverSetup.String())

		if serverSetup.SelectedVersion != wire.DRAFT_04 {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "SERVER SETUP ERRR : PROTOCOL DRAFT NOT SUPPORTED")
			return
		}

		if role, err = serverSetup.GetRoleParam(); err != nil {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "CLIENT SETUP ERROR : ROLE PARAM MISSING")
			return
		}

		if err := cs.SetRemoteRole(role); err != nil {
			cs.Close(wire.MOQERR_PROTOCOL_VIOLATION, err.Error())
			return
		}

		cs.ishandshakedone = true
		cs.HandshakeDone <- true
		cs.Slogger.Info().Msgf("[Handshake Success]")

	default:
		log.Error().Msgf("[Received Unknown Setup Message][Type - %s][%X]", wire.GetMoqMessageString(m.Type()), m.Type())
	}
}

func (cs *ControlStream) handleControlMessage(m wire.MOQTMessage) {

	switch m.Type() {
	case wire.ANNOUNCE:
		cs.Handler.HandleAnnounce(m.(*wire.Announce))
	case wire.SUBSCRIBE:
		cs.Handler.HandleSubscribe(m.(*wire.Subscribe))
	case wire.SUBSCRIBE_OK:
		cs.Handler.HandleSubscribeOk(m.(*wire.SubscribeOk))
	case wire.ANNOUNCE_OK:
		cs.Handler.HandleAnnounceOk(m.(*wire.AnnounceOk))
	case wire.UNSUBSCRIBE:
		cs.Handler.HandleUnsubscribe(m.(*wire.Unsubcribe))
	default:
		log.Error().Msgf("Unknown Control Message +v", m)
	}
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

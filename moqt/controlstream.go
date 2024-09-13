package moqt

import (
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

	cs.Slogger.Debug().Msgf("[Dipsatching CONTROL]%s", msg.String())
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

	switch m.Type() {
	case wire.CLIENT_SETUP:

		clientSetup := m.(*wire.ClientSetup)

		cs.Slogger.Debug().Msgf(clientSetup.String())

		if !clientSetup.CheckDraftSupport() {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "CLIENT SETUP ERROR : PROTOCOL DRAFT NOT SUPPORTED")
			return
		}

		if role := clientSetup.Params.GetParameter(wire.ROLE_PARAM); role != nil {
			cs.SetRemoteRole(role.Value().(uint64))
		} else {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "CLIENT SETUP ERROR : ROLE PARAM MISSING")
			return
		}

		cs.ishandshakedone = true
		cs.Slogger.Info().Msgf("[Handshake Success]")
		cs.WriteControlMessage(&DEFAULT_SERVER_SETUP)

	case wire.SERVER_SETUP:

		serverSetup := m.(*wire.ServerSetup)

		cs.Slogger.Debug().Msgf(serverSetup.String())

		if serverSetup.SelectedVersion != wire.DRAFT_04 {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "SERVER SETUP ERRR : PROTOCOL DRAFT NOT SUPPORTED")
			return
		}

		if role := serverSetup.Params.GetParameter(wire.ROLE_PARAM); role != nil {
			cs.SetRemoteRole(role.Value().(uint64))
		} else {
			cs.Close(wire.MOQERR_INTERNAL_ERROR, "SERVER SETUP ERROR : ROLE PARAM MISSING")
			return
		}

		cs.ishandshakedone = true
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

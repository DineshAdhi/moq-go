package moqt

import (
	"moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
)

type RelayHandler struct {
	*MOQTSession
}

func (handler *RelayHandler) HandleAnnounce(msg *wire.AnnounceMessage) {

	handler.Slogger.Info().Msgf(msg.String())

	okmsg := wire.AnnounceOkMessage{}
	okmsg.TrackNameSpace = msg.TrackNameSpace

	sm.addPublisher(msg.TrackNameSpace, handler.MOQTSession)

	handler.CS.WriteControlMessage(&okmsg)
}

func (handler *RelayHandler) HandleSubscribe(msg *wire.SubscribeMessage) {

	handler.Slogger.Info().Msg(msg.String())

	pub := sm.getPublisher(msg.ObjectStreamNamespace)

	if pub == nil {
		log.Error().Msgf("[No Publisher found with Namespace - %s]", msg.ObjectStreamNamespace)
		return
	}

	os := pub.GetObjectStream(msg)
	os.AddSubscriber(msg.SubscribeID, handler.MOQTSession)

	handler.SubscribedStreams.AddStream(msg.SubscribeID, os)
}

func (handler *RelayHandler) HandleSubscribeOk(msg *wire.SubscribeOkMessage) {
	handler.Slogger.Info().Msg(msg.String())
}

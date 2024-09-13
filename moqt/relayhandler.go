package moqt

import (
	"moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
)

type RelayHandler struct {
	*MOQTSession
}

// Comes from Publisher
func (publisher *RelayHandler) HandleAnnounce(msg *wire.Announce) {

	publisher.Slogger.Info().Msgf(msg.String())

	okmsg := wire.AnnounceOk{}
	okmsg.TrackNameSpace = msg.TrackNameSpace

	sm.addPublisher(msg.TrackNameSpace, publisher.MOQTSession)

	publisher.CS.WriteControlMessage(&okmsg)
}

// Comes from Publisher
func (publisher *RelayHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {
	publisher.Slogger.Info().Msg(msg.String())

	subid := msg.SubscribeID

	if os, ok := publisher.IncomingStreams.SubIDGetStream(subid); ok {
		for _, sub := range os.subscribers {
			sub.CS.SendSubscribeOk(os.streamid, msg)
		}
	}
}

func (publisher *RelayHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {
	publisher.Slogger.Info().Msg(msg.String())
}

// Comes from Subscriber
func (subscriber *RelayHandler) HandleSubscribe(msg *wire.Subscribe) {

	subscriber.Slogger.Info().Msg(msg.String())

	pub := sm.getPublisher(msg.TrackNameSpace)

	if pub == nil {
		log.Error().Msgf("[No Publisher found with Namespace - %s]", msg.TrackNameSpace)
		return
	}

	os := pub.GetObjectStream(msg)
	os.AddSubscriber(msg.SubscribeID, subscriber.MOQTSession)

	subscriber.SubscribedStreams.AddStream(msg.SubscribeID, os)
}

func (subscriber *RelayHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	subscriber.Slogger.Info().Msg(msg.String())
}

func (subscriber *RelayHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

	subscriber.Slogger.Info().Msg(msg.String())

	subid := msg.SubscriptionID

	if os, ok := subscriber.SubscribedStreams.SubIDGetStream(subid); ok {
		os.RemoveSubscriber(subscriber.id)
	}
}

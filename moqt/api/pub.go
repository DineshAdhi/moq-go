package api

import (
	"context"
	"moq-go/moqt"
	"moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
)

const catalog string = "{\"version\":1,\"streamingFormat\":1,\"streamingFormatVersion\":\"0.2\",\"supportsDeltaUpdates\":true,\"commonTrackFields\":{\"namespace\":\"bbb\",\"packaging\":\"cmaf\",\"renderGroup\":1},\"tracks\":[{\"name\":\"1.m4s\",\"initTrack\":\"0.mp4\",\"selectionParams\":{\"codec\":\"avc1.64001F\",\"width\":1280,\"height\":720},\"namespace\":\"bbb\",\"packaging\":\"cmaf\",\"renderGroup\":1},{\"name\":\"2.m4s\",\"initTrack\":\"0.mp4\",\"selectionParams\":{\"codec\":\"mp4a.40.2\",\"bitrate\":125587,\"samplerate\":44100,\"channelConfig\":\"2\"},\"namespace\":\"bbb\",\"packaging\":\"cmaf\",\"renderGroup\":1}]}"

type Publisher struct {
	Options          moqt.DialerOptions
	Relay            string
	namespace        string
	Ctx              context.Context
	SubscriptionChan chan *moqt.PubStream
}

func NewMOQTPublisher(options moqt.DialerOptions, ns string, relay string) *Publisher {

	pub := &Publisher{
		Options:          options,
		Relay:            relay,
		namespace:        ns,
		Ctx:              context.Background(),
		SubscriptionChan: make(chan *moqt.PubStream),
	}

	return pub
}

func (pub *Publisher) Run() error {

	dialer := moqt.MOQTDialer{
		Options: pub.Options,
		Ctx:     pub.Ctx,
		Role:    wire.ROLE_PUBLISHER,
	}

	session, err := dialer.Dial(pub.Relay)

	if err != nil {
		log.Error().Msgf("[Failed to connect to Relay][%s]", err)
		return err
	}

	handler := session.Handler.(*moqt.PubHandler)
	handler.Announce(pub.namespace)

	go func() {
		for {
			stream := <-handler.SubscribeChannel

			if stream.TrackName == "counter" {
				handler.SubscribeOk(stream)
				pub.SubscriptionChan <- stream
			}
		}
	}()

	return nil
}

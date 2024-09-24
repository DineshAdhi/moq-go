package api

import (
	"context"
	"moq-go/moqt"
	"moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
)

type Subscriber struct {
	session          *moqt.MOQTSession
	Options          moqt.DialerOptions
	Relay            string
	namespace        string
	Ctx              context.Context
	SubscriptionChan chan *moqt.SubStream
}

func NewMOQTSubscriber(options moqt.DialerOptions, ns string, relay string) *Subscriber {

	sub := &Subscriber{
		Options:          options,
		Relay:            relay,
		namespace:        ns,
		Ctx:              context.Background(),
		SubscriptionChan: make(chan *moqt.SubStream),
	}

	return sub
}

func (sub *Subscriber) Run() error {

	var err error

	dialer := moqt.MOQTDialer{
		Options: sub.Options,
		Ctx:     sub.Ctx,
		Role:    wire.ROLE_SUBSCRIBER,
	}

	sub.session, err = dialer.Dial(sub.Relay)

	if err != nil {
		log.Error().Msgf("[Failed to connect to Relay][%s]", err)
		return err
	}

	handler := sub.session.Handler.(*moqt.SubHandler)

	go func() {
		for {
			ss := <-handler.StreamsChan
			sub.SubscriptionChan <- ss
		}
	}()

	return nil
}

func (sub *Subscriber) Subscribe(name string, alias uint64) {
	handler := sub.session.Handler.(*moqt.SubHandler)
	handler.Subscribe(sub.namespace, name, 0)
}

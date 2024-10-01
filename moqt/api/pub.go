package api

import (
	"context"
	"moq-go/moqt"
	"moq-go/moqt/wire"
)

type MOQPub struct {
	Options            moqt.DialerOptions
	Ctx                context.Context
	Relay              string
	handler            *moqt.PubHandler
	onSubscribeHandler func(moqt.PubStream)
}

func NewMOQPub(options moqt.DialerOptions, relay string) *MOQPub {
	pub := &MOQPub{
		Options: options,
		Ctx:     context.TODO(),
		Relay:   relay,
	}

	return pub
}

func (pub *MOQPub) OnSubscribe(f func(moqt.PubStream)) {
	pub.onSubscribeHandler = f
}

func (pub *MOQPub) Connect() (*moqt.PubHandler, error) {

	dialer := moqt.MOQTDialer{
		Options: pub.Options,
		Role:    wire.ROLE_PUBLISHER,
		Ctx:     pub.Ctx,
	}

	session, err := dialer.Dial(pub.Relay)

	if err != nil {
		return nil, err
	}

	pub.handler = session.PubHandler()

	go func() {
		for stream := range pub.handler.SubscribeChan {
			pub.onSubscribeHandler(stream)
		}
	}()

	return session.PubHandler(), nil
}

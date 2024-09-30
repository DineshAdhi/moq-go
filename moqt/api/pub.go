package api

import (
	"context"
	"moq-go/moqt"
	"moq-go/moqt/wire"
)

type MOQPub struct {
	Options moqt.DialerOptions
	Ctx     context.Context
	Relay   string
}

func NewMOQPub(options moqt.DialerOptions, relay string) *MOQPub {
	pub := &MOQPub{
		Options: options,
		Ctx:     context.TODO(),
		Relay:   relay,
	}
	return pub
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

	return session.PubHandler(), nil
}

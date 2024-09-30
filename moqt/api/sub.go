package api

import (
	"context"
	"moq-go/moqt"
	"moq-go/moqt/wire"
)

type MOQSub struct {
	Options moqt.DialerOptions
	Relay   string
	Ctx     context.Context
}

func NewMOQSub(options moqt.DialerOptions, relay string) *MOQSub {
	sub := &MOQSub{
		Options: options,
		Relay:   relay,
		Ctx:     context.TODO(),
	}

	return sub
}

func (pub *MOQSub) Connect() (*moqt.SubHandler, error) {

	dialer := moqt.MOQTDialer{
		Options: pub.Options,
		Role:    wire.ROLE_SUBSCRIBER,
		Ctx:     pub.Ctx,
	}

	session, err := dialer.Dial(pub.Relay)

	if err != nil {
		return nil, err
	}

	return session.SubHandler(), nil
}

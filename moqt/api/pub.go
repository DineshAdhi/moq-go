package api

import (
	"context"
	"moq-go/moqt"
	"moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
)

type Publisher struct {
	Options moqt.DialerOptions
	Relays  []string
	Ctx     context.Context
}

func NewMOQTPublisher(options moqt.DialerOptions, relays []string) *Publisher {

	pub := &Publisher{
		Options: options,
		Relays:  relays,
		Ctx:     context.Background(),
	}

	return pub
}

func (pub *Publisher) Run() error {

	dialer := moqt.MOQTDialer{
		Options: pub.Options,
		Ctx:     pub.Ctx,
		Role:    wire.ROLE_PUBLISHER,
	}

	session, err := dialer.Dial("127.0.0.1:4443")

	if err != nil {
		log.Error().Msgf("[Relay Listen Failed][%s]", err)
	}

	go session.ServeMOQ()

	return nil
}

package api

import (
	"context"

	"github.com/DineshAdhi/moq-go/moqt"
	"github.com/DineshAdhi/moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
)

type MOQRelay struct {
	Options moqt.ListenerOptions
	Peers   []string // StringArray containing PeerAddress
	Ctx     context.Context
}

func NewMOQTRelay(options moqt.ListenerOptions, peers []string) *MOQRelay {

	relay := &MOQRelay{
		Options: options,
		Peers:   peers,
		Ctx:     context.TODO(),
	}

	return relay
}

func (relay *MOQRelay) Run() error {

	listener := moqt.MOQTListener{
		Options: relay.Options,
		Ctx:     relay.Ctx,
		Role:    wire.ROLE_RELAY,
	}

	err := listener.Listen()
	log.Error().Msgf("[Relay Listen Failed][%s]", err)
	relay.Ctx.Done()

	return nil
}

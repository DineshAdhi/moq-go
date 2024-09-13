package api

import (
	"context"
	"moq-go/moqt"
	"moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
)

type Relay struct {
	Options moqt.ListenerOptions
	Role    uint64
	Peers   []string // StringArray containing PeerAddress
	Ctx     context.Context
}

func NewMOQTRelay(options moqt.ListenerOptions, peers []string) *Relay {

	relay := &Relay{
		Options: options,
		Role:    wire.ROLE_RELAY,
		Peers:   peers,
		Ctx:     context.TODO(),
	}

	return relay
}

func (relay *Relay) Run() error {

	listener := moqt.MOQTListener{
		Options: relay.Options,
		Ctx:     relay.Ctx,
		Role:    relay.Role,
	}

	err := listener.Listen()
	log.Error().Msgf("[Relay Listen Failed][%s]", err)
	relay.Ctx.Done()

	return nil
}

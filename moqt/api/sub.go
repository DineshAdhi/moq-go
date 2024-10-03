package api

import (
	"context"

	"github.com/DineshAdhi/moq-go/moqt"
	"github.com/DineshAdhi/moq-go/moqt/wire"
)

type MOQSub struct {
	Options           moqt.DialerOptions
	Relay             string
	Ctx               context.Context
	onStreamHandler   func(moqt.SubStream)
	onAnnounceHandler func(string)
	handler           *moqt.SubHandler
}

func NewMOQSub(options moqt.DialerOptions, relay string) *MOQSub {
	sub := &MOQSub{
		Options: options,
		Relay:   relay,
		Ctx:     context.TODO(),
	}

	return sub
}

func (sub *MOQSub) OnStream(f func(moqt.SubStream)) {
	sub.onStreamHandler = f
}

func (sub *MOQSub) OnAnnounce(f func(string)) {
	sub.onAnnounceHandler = f
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

	handler := session.SubHandler()

	go func() {
		for stream := range handler.StreamsChan {
			pub.onStreamHandler(stream)
		}
	}()

	go func() {
		for ns := range handler.AnnounceChan {
			pub.onAnnounceHandler(ns)
		}
	}()

	return session.SubHandler(), nil
}

package main

import (
	"flag"
	"moq-go/moqt"
	"moq-go/moqt/api"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var ALPNS = []string{"moq-00"} // Application Layer Protocols ["H3" - WebTransport]
const RELAY = "127.0.0.1:4443"

func main() {
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	Options := moqt.DialerOptions{
		ALPNs: ALPNS,
		QuicConfig: &quic.Config{
			EnableDatagrams: true,
		},
	}

	sub := api.NewMOQTSubscriber(Options, "bbb", RELAY)
	sub.Run()

	sub.Subscribe("2.m4s", 0)

	go func() {
		for {
			ss := <-sub.SubscriptionChan
			go handleStream(ss)
		}
	}()
}

func handleStream(ss *moqt.SubStream) {
	for {
		toject := <-ss.ObjectChan
		data := string(toject.Payload[:])

		log.Debug().Msgf("Sub Data : %s", data)
	}
}

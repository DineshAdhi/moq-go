package main

import (
	"context"
	"flag"
	moqt "moq-go/moqt"
	"moq-go/moqt/api"
	"moq-go/moqt/wire"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const LISTENADDR = "0.0.0.0:4443"

const CERTPATH = "../certs/localhost.crt"
const KEYPATH = "../certs/localhost.key"

var ALPNS = []string{"moq-00"} // Application Layer Protocols ["H3" - WebTransport]

func main() {

	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	Options := moqt.ListenerOptions{
		ListenAddr: "0.0.0.0:5553",
		CertPath:   CERTPATH,
		KeyPath:    KEYPATH,
		ALPNs:      []string{"moq-00", "h3"},
		QuicConfig: nil,
	}

	relay := api.NewMOQTRelay(Options, []string{})

	DialerOptions := moqt.DialerOptions{
		DialAddress: "127.0.0.1:4443",
		CertPath:    CERTPATH,
		KeyPath:     KEYPATH,
		ALPNs:       ALPNS,
		QuicConfig:  nil,
	}

	dialer := moqt.MOQTDialer{
		Options: DialerOptions,
		Role:    wire.ROLE_RELAY,
		Ctx:     context.TODO(),
	}

	session, err := dialer.Connect()

	if err != nil {
		log.Error().Msgf("[Error Connect][%+v]", err)
	}

	go session.ServeMOQ()

	relay.Run()
}

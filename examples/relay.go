package main

import (
	"context"
	"flag"
	moqt "moq-go/moqt"
	"moq-go/moqt/wire"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const LISTENADDR = "0.0.0.0:4443"

const CERTPATH = "./certs/localhost.crt"
const KEYPATH = "./certs/localhost.key"

var ALPNS = []string{"h3", "moq-00"} // Application Layer Protocols ["H3" - WebTransport]

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener := moqt.MOQTListener{
		ListenAddr: LISTENADDR,
		CertPath:   CERTPATH,
		KeyPath:    KEYPATH,
		ALPNS:      ALPNS,
		QuicConfig: nil,
		Ctx:        ctx,
		Role:       wire.ROLE_RELAY,
	}

	err := listener.Listen()

	log.Error().Msgf("[Error MOQListener][%s]", err)
}

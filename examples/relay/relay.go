package main

import (
	"flag"
	"fmt"
	moqt "moq-go/moqt"
	"moq-go/moqt/api"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const LISTENADDR = "0.0.0.0:4443"

var ALPNS = []string{"h3", "moq-00"} // Application Layer Protocols ["H3" - WebTransport]

func main() {

	debug := flag.Bool("debug", false, "sets log level to debug")
	port := flag.Int("port", 4443, "Listening Port")
	KEYPATH := flag.String("keypath", "../certs/localhost.key", "Keypath")
	CERTPATH := flag.String("certpath", "../certs/localhost.crt", "CertPath")
	flag.Parse()

	LISTENADDR := fmt.Sprintf("0.0.0.0:%d", *port)

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	Options := moqt.ListenerOptions{
		ListenAddr: LISTENADDR,
		CertPath:   *CERTPATH,
		KeyPath:    *KEYPATH,
		ALPNs:      ALPNS,
		QuicConfig: nil,
	}

	peers := []string{} // TODO : Address of the Relay Peers for Fan out Implementation

	relay := api.NewMOQTRelay(Options, peers)
	relay.Run()
}

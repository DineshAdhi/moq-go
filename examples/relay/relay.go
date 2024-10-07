package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DineshAdhi/moq-go/moqt"

	"github.com/DineshAdhi/moq-go/moqt/api"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const PORT = 4443

var ALPNS = []string{"h3", "moq-00"} // Application Layer Protocols ["H3" - WebTransport]

func main() {
	// defer profile.Start(profile.ProfilePath("."), profile.GoroutineProfile, profile.MemProfileHeap, profile.CPUProfile).Stop()

	ENVCERTPATH := os.Getenv("MOQT_CERT_PATH")
	ENVKEYPATH := os.Getenv("MOQT_KEY_PATH")

	debug := flag.Bool("debug", false, "sets log level to debug")
	port := flag.Int("port", PORT, "Listening Port")
	KEYPATH := flag.String("keypath", ENVKEYPATH, "Keypath")
	CERTPATH := flag.String("certpath", ENVCERTPATH, "CertPath")
	flag.Parse()

	LISTENADDR := fmt.Sprintf("0.0.0.0:%d", *port)

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	quicConfig := &quic.Config{
		EnableDatagrams: true,
	}

	Options := moqt.ListenerOptions{
		ListenAddr: LISTENADDR,
		CertPath:   *CERTPATH,
		KeyPath:    *KEYPATH,
		ALPNs:      ALPNS,
		QuicConfig: quicConfig,
	}

	peers := []string{} // TODO : Address of the Relay Peers for Fan out Implementation

	relay := api.NewMOQTRelay(Options, peers)
	relay.Run()
}

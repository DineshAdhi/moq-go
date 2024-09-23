package main

import (
	"context"
	"flag"
	"moq-go/moqt"
	"moq-go/moqt/wire"
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

func main() {
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	ENVCERTPATH := os.Getenv("MOQT_CERT_PATH")
	ENVKEYPATH := os.Getenv("MOQT_KEY_PATH")

	debug := flag.Bool("debug", false, "sets log level to debug")
	KEYPATH := flag.String("keypath", ENVKEYPATH, "Keypath")
	CERTPATH := flag.String("certpath", ENVCERTPATH, "CertPath")
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
		CertPath: *CERTPATH,
		KeyPath:  *KEYPATH,
		ALPNs:    ALPNS,
		QuicConfig: &quic.Config{
			EnableDatagrams: true,
		},
	}

	dialer := moqt.MOQTDialer{
		Options: Options,
		Role:    wire.ROLE_PUBLISHER,
		Ctx:     context.Background(),
	}

	session, err := dialer.Dial("127.0.0.1:4443")

	if err != nil {
		log.Error().Msgf("[Error Connecting to Relay][%s]", err)
		return
	}

	submsg := wire.Subscribe{
		SubscribeID:    1,
		TrackAlias:     0,
		TrackNameSpace: "bbb",
		TrackName:      "dumeel",
		FilterType:     wire.LatestGroup,
	}

	session.SendSubscribe(&submsg)

	for {

	}
}

package main

import (
	"flag"
	"io"
	"moq-go/moqt"
	"moq-go/moqt/api"
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

	sub := api.NewMOQSub(Options, RELAY)
	handler, err := sub.Connect()

	handler.Subscribe("bbb", "dumeel", 0)

	for substream := range handler.SubscribedChan {
		go handleStream(&substream)
	}

	if err != nil {
		log.Error().Msgf("Error - %s", err)
		return
	}
}

func handleStream(ss *moqt.SubStream) {

	log.Debug().Msgf("New Stream Header")

	for moqtstream := range ss.StreamsChan {
		log.Debug().Msgf("New Group Stream - %s", moqtstream.GetStreamID())
		go handleMOQStream(moqtstream)
	}
}

func handleMOQStream(stream wire.MOQTStream) {

	for {
		groupid, object, err := stream.ReadObject()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Error().Msgf("Error Reading Objects - %s", err)
			break
		}

		msg := string(object.Payload[:])
		log.Printf("Payload - %d %s - %d", groupid, msg, object.ID)
	}

	log.Printf("Group Ended")

}

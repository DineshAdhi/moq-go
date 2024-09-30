package main

import (
	"flag"
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

const PORT = 4443

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

	pub := api.NewMOQPub(Options, RELAY)
	handler, err := pub.Connect()

	if err != nil {
		log.Error().Msgf("error - %s", err)
		return
	}

	handler.SendAnnounce("bbb")

	for pubstream := range handler.StreamsChan {
		go handleStream(&pubstream)
	}
}

func handleStream(stream *moqt.PubStream) {
	stream.Accept()

	groupid := 0

	for {
		gs, err := stream.NewTrack()

		if err != nil {
			log.Error().Msgf("Err - %s", err)
			return
		}

		objectid := 0

		for range 5 {

			obj := &wire.Object{
				GroupID: uint64(groupid),
				ID:      uint64(objectid),
				Payload: []byte("Dinesh"),
			}

			gs.WriteObject(obj)
			objectid++
		}

		gs.Close()

		groupid++

		<-time.After(time.Millisecond * 250)
	}
}

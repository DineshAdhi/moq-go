package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DineshAdhi/moq-go/moqt"
	"github.com/DineshAdhi/moq-go/moqt/api"
	"github.com/DineshAdhi/moq-go/moqt/wire"
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

	pub.OnSubscribe(func(ps moqt.PubStream) {
		log.Debug().Msgf("New Subscribe Request - %s", ps.TrackName)
		go handleStream(&ps)
	})

	if err != nil {
		log.Error().Msgf("error - %s", err)
		return
	}

	handler.SendAnnounce("bbb")

	<-pub.Ctx.Done()
}

func handleStream(stream *moqt.PubStream) {
	stream.Accept()

	groupid := uint64(0)

	for {
		gs, err := stream.NewGroup(groupid)

		if err != nil {
			log.Error().Msgf("Err - %s", err)
			return
		}

		objectid := uint64(0)

		for range 10 {
			gs.WriteObject(&wire.Object{
				GroupID: groupid,
				ID:      objectid,
				Payload: []byte("Dinesh"),
			})
			objectid++
		}

		gs.Close()

		groupid++

		<-time.After(time.Millisecond * 50)
	}
}

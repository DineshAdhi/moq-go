package main

import (
	"flag"
	"io"
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

var ALPNS = []string{"moq-00"} // Application Layer Protocols ["H3" - WebTransport]
const RELAY = "127.0.0.1:4443"

var GROUPID uint64 = 0

func main() {

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

	sub.OnStream(func(ss moqt.SubStream) {
		go handleStream(&ss)
	})

	sub.OnAnnounce(func(ns string) {
		// handler.Subscribe(ns, "dumeel", 0)
	})

	if err != nil {
		log.Error().Msgf("Error - %s", err)
		return
	}

	handler.Subscribe("bbb", "dumeel", 0)

	<-sub.Ctx.Done()
}

func handleStream(ss *moqt.SubStream) {
	for moqtstream := range ss.StreamsChan {
		handleMOQStream(moqtstream)
	}
}

func handleMOQStream(stream wire.MOQTStream) {

	gs := stream.(*wire.GroupStream)

	objectcount := 0
	var arr []uint64

	for {
		_, object, err := stream.ReadObject()

		if err == io.EOF {
			break
		}

		objectcount++

		if err != nil {
			log.Error().Msgf("Error Reading Objects - %s", err)
			break
		}

		arr = append(arr, object.ID)
	}

	if GROUPID != gs.GroupID {
		log.Error().Msgf("Received Wrong GroupID - Expected %d, Got - %d", GROUPID, gs.GroupID)
		GROUPID = gs.GroupID
	}

	GROUPID++

	if objectcount != 10 {
		log.Error().Msgf("Incomplete Group : %d %+v", gs.GroupID, arr)
	} else {
		// log.Info().Msgf("New Group [%d] - %+v", gs.GroupID, arr)
	}
}

package main

import (
	"flag"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
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
const RELAY = "localhost:4443"

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
		// handler.Subscribe(ns, "teststream", 0)
	})

	if err != nil {
		log.Error().Msgf("Error - %s", err)
		return
	}

	handler.Subscribe("ffmpegtest", "teststream", 0)

	<-sub.Ctx.Done()
}

func handleStream(ss *moqt.SubStream) {

	gsMap := make(map[uint64]*wire.GroupStream)

	for moqtstream := range ss.StreamsChan {
		gs := moqtstream.(*wire.GroupStream)
		gsMap[gs.GroupID] = gs

		if len(gsMap) >= 25 {

			keys := make([]uint64, 0, len(gsMap))
			for k := range gsMap {
				keys = append(keys, k)
			}

			sort.Slice(keys, func(i, j int) bool {
				return keys[i] < keys[j]
			})

			log.Info().Msgf("Group ID Order %+v", keys)

			oldkey := -1

			for _, key := range keys {

				if oldkey == -1 {
					oldkey = int(key)
				} else if oldkey != int(key) {
					break
				}

				handleMOQStream(gsMap[key])
				delete(gsMap, key)

				oldkey++
			}
		}
	}
}

func handleMOQStream(gs *wire.GroupStream) {

	conn, err := net.Dial("udp", "127.0.0.1:4000")

	if err != nil {
		panic(err)
	}

	for {
		_, object, err := gs.ReadObject()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Error().Msgf("Error Reading Objects - %s", err)
			break
		}

		conn.Write(object.Payload)
	}

	if GROUPID != gs.GroupID {
		log.Error().Msgf("Received Wrong GroupID - Expected %d, Got - %d", GROUPID, gs.GroupID)
		GROUPID = gs.GroupID
	}

	GROUPID++
}

package main

import (
	"bufio"
	"flag"
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

func ReadFromStdin() chan []byte {

	datachannel := make(chan []byte)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			var data [2048]byte
			n, err := reader.Read(data[:])

			if err != nil {
				break
			}

			datachannel <- data[:n]
		}
	}()

	return datachannel
}

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
			EnableDatagrams:       true,
			MaxIncomingUniStreams: 1000,
		},
	}

	pub := api.NewMOQPub(Options, RELAY)
	handler, err := pub.Connect()

	pub.OnSubscribe(func(ps moqt.PubStream) {
		go handleStream(&ps)
	})

	if err != nil {
		log.Error().Msgf("error - %s", err)
		return
	}

	handler.SendAnnounce("ffmpegtest")

	<-pub.Ctx.Done()
}

func handleStream(stream *moqt.PubStream) {
	stream.Accept()

	ch := ReadFromStdin()

	groupid := uint64(0)
	objid := uint64(0)

	for {
		gs, err := stream.NewGroup(groupid)

		if err != nil {
			log.Error().Msgf("%s", err)
			break
		}

		for range 1000 {
			data := <-ch

			obj := &wire.Object{
				ID:      objid,
				Payload: data,
			}

			gs.WriteObject(obj)
			objid++
		}

		gs.Close()

		groupid++
	}
}

package main

import (
	"flag"
	"fmt"
	"moq-go/moqt"
	"moq-go/moqt/api"
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

	pub := api.NewMOQTPublisher(Options, "bbb", RELAY)
	pub.Run()

	for {
		stream := <-pub.SubscriptionChan
		go handleStream(stream)
	}
}

func handleStream(stream *moqt.PubStream) {

	log.Debug().Msgf("Handing Stream  : %s", stream.StreamId)

	var itr uint64 = 0

	for {
		str := fmt.Sprintf("Dinesh %d", itr)
		data := []byte(str)
		stream.Push(itr, data)
		itr++
		<-time.After(time.Second)
	}

	// data := [6]byte{}

	// for {
	// 	n, err := os.Stdin.Read(data[:])
	// 	if err != nil {
	// 		stream.Push(itr, data[:n])
	// 		log.Debug().Msgf("Pusing : %s", string(data[:n]))
	// 		itr++
	// 	}
	// }
}

package main

import (
	"moq-go/logger"
	"moq-go/moqt"
	"moq-go/wt"
	"net/http"
)

const LISTENADDR = "0.0.0.0:4443"

const CERTPATH = "./certs/localhost.crt"
const KEYPATH = "./certs/localhost.key"

var ALPNS = []string{"h3", "moq-00"} // Application Layer Protocols ["H3" - WebTransport]

func main() {

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		wts := req.Body.(*wt.WTSession)
		wts.AcceptSession()

		moqtsession := moqt.MOQTSession{
			Conn: wts,
		}

		moqtsession.Serve()
	})

	wtserver := wt.WTServer{
		ListenAddr: LISTENADDR,
		CertPath:   CERTPATH,
		KeyPath:    KEYPATH,
		ALPNS:      ALPNS,
		QuicConfig: nil,
	}

	err := wtserver.Run()

	logger.ErrorLog("[WTS Error][%s]", err)
}

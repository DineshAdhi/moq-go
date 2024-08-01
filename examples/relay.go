package main

import (
	"log"
	"moq-go/wt"
	"net/http"
)

const LISTENADDR = "0.0.0.0:4443"

const CERTPATH = "./certs/localhost.crt"
const KEYPATH = "./certs/localhost.key"

var ALPNS = []string{"h3"} // Application Layer Protocols ["H3" - WebTransport]

func main() {

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		wts := req.Body.(*wt.WTSession)
		wts.AcceptSession()

		_, err := wts.AcceptStream()

		if err != nil {
			log.Printf("[Error Accepting Stream from WTS]%s", err)
			return
		}
	})

	wtserver := wt.WTServer{
		ListenAddr: LISTENADDR,
		CertPath:   CERTPATH,
		KeyPath:    KEYPATH,
		ALPNS:      ALPNS,
		QuicConfig: nil,
	}

	err := wtserver.Run()

	log.Printf("[WTS Error][%s]", err)

}

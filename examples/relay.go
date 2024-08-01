package main

import (
	"log"
	"moq-go/moqt"
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

		bistream, err := wts.AcceptStream()

		if err != nil {
			log.Printf("[Error Accepting Stream from WTS]%s", err)
			return
		}

		clientsetup := &moqt.ClientSetup{}
		err = clientsetup.Read(bistream)

		if err != nil {
			log.Printf("[Error Receiving Client Setup]%s", err)
			return
		}

		log.Printf("[Client Setup][%+v]", clientsetup)

		serverSetup := moq.DefaultServerSetup()
		bistream.Write(serverSetup.GetBytes())

		log.Printf("[Sent SERVER SETUP][%+v]", serverSetup)

		for {
			msg := moqt.MOQTMessage{}
			err := msg.Read(bistream)

			if err != nil {
				log.Printf("[Error Reading MOQT Message]%s", err)
				break
			}

			log.Printf("[Got MOQT Mesage][%s]", moqt.GetMoqMessageString(msg.Mtype))
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

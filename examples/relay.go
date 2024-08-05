package main

import (
	"log"
	"moq-go/moqt"
	"moq-go/wt"
	"net/http"

	"github.com/quic-go/quic-go/quicvarint"
)

const LISTENADDR = "0.0.0.0:4443"

const CERTPATH = "./certs/localhost.crt"
const KEYPATH = "./certs/localhost.key"

var ALPNS = []string{"h3"} // Application Layer Protocols ["H3" - WebTransport]

var DEFAULT_SERVER_SETUP = moqt.ServerSetup{SelectedVersion: moqt.DRAFT_03, Params: moqt.Parameters{
	moqt.ROLE_PARAM: &moqt.IntParameter{moqt.ROLE_PARAM, moqt.ROLE_PUBSUB},
}}

func main() {

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		wts := req.Body.(*wt.WTSession)
		wts.AcceptSession()

		controlStream, err := wts.AcceptStream()
		controlReader := quicvarint.NewReader(controlStream)

		if err != nil {
			log.Printf("[Error Accepting Stream from WTS]%s", err)
			return
		}

		// 1. Client Setup Parsing

		mtype, msg, err := moqt.ParseMOQTMessage(controlReader)

		if err != nil {
			log.Printf("[Error Receiving MOQT Message][%s]", err)
			return
		}

		if mtype != moqt.CLIENT_SETUP {
			log.Printf("[Client Setup Error][Unexpected MOQT Message][Received - %s(%X)]", moqt.GetMoqMessageString(mtype), mtype)
			return
		}

		clientSetup := msg.(*moqt.ClientSetup)

		if !clientSetup.CheckDraftSupport() {
			log.Printf("[Client Setup Error][Unsupported Draft Versions][%+v]", clientSetup.SupportedVersions)
			return
		}

		log.Printf("[Received Client Setup][%s]", clientSetup.String())

		// 2. Server Setup Dispatching

		serverSetup := DEFAULT_SERVER_SETUP
		_, err = controlStream.Write(serverSetup.GetBytes())

		if err != nil {
			log.Printf("[Server Setup Dispatch Error][%s]", err)
			return
		}

		log.Printf("[Sent Server Setup][%s]", serverSetup.String())

		// 3. Wait for Announce

		mtype, msg, err = moqt.ParseMOQTMessage(controlReader)

		log.Printf("%s", msg.String())

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

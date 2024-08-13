package main

import (
	"context"
	"moq-go/logger"
	"moq-go/moqt"
)

const LISTENADDR = "0.0.0.0:4443"

const CERTPATH = "./certs/localhost.crt"
const KEYPATH = "./certs/localhost.key"

var ALPNS = []string{"h3", "moq-00"} // Application Layer Protocols ["H3" - WebTransport]

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener := moqt.MOQTListener{
		ListenAddr: LISTENADDR,
		CertPath:   CERTPATH,
		KeyPath:    KEYPATH,
		ALPNS:      ALPNS,
		QuicConfig: nil,
		Ctx:        ctx,
	}

	err := listener.Listen()

	logger.ErrorLog("[Error MOQListener][%s]", err)
}

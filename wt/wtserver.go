package wt

import (
	"context"
	"crypto/tls"
	"moq-go/logger"
	"net/http"

	"github.com/quic-go/quic-go"
)

type WTServer struct {
	http.Handler
	ListenAddr string
	CertPath   string
	KeyPath    string
	ALPNS      []string
	QuicConfig *quic.Config
}

// WebTransport Server Implementation

func (wtserver *WTServer) Run() error {

	if wtserver.Handler == nil {
		wtserver.Handler = http.DefaultServeMux
	}

	tlsCerts, err := tls.LoadX509KeyPair(wtserver.CertPath, wtserver.KeyPath)

	if err != nil {
		return err
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{tlsCerts},
		NextProtos:   wtserver.ALPNS,
	}

	quicListener, err := quic.ListenAddr(wtserver.ListenAddr, &tlsConfig, wtserver.QuicConfig)

	if err != nil {
		return err
	}

	logger.InfoLog("[WebTransport Server][Listening on - %s]", wtserver.ListenAddr)

	for {
		quicConn, err := quicListener.Accept(context.Background())

		if err != nil {
			return err
		}

		go wtserver.handleConnection(quicConn)
	}
}

func (server *WTServer) handleConnection(quicConn quic.Connection) {

	wts, req, err := UpgradeWTS(quicConn)

	if err != nil {
		logger.ErrorLog("[Error Upgrading WT Session to MOQ][%s]", err)
		return
	}

	server.ServeHTTP(wts.ResponseWriter, req)
}

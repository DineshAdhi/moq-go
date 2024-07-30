package wt

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	"github.com/quic-go/quic-go"
)

type WTServer struct {
	http.Handler
	ListenAddr string
	CertPath   string
	KeyPath    string
	ALPNs      []string
}

func (server *WTServer) Run() error {

	if server.Handler == nil {
		server.Handler = http.DefaultServeMux
	}

	tlscert, err := tls.LoadX509KeyPair(server.CertPath, server.KeyPath)

	if err != nil {
		return err
	}

	tlsconfig := &tls.Config{
		Certificates: []tls.Certificate{tlscert},
		NextProtos:   server.ALPNs,
	}

	listener, err := quic.ListenAddr(server.ListenAddr, tlsconfig, nil)

	if err != nil {
		return err
	}

	log.Printf("Listening on : %s", server.ListenAddr)

	for {
		conn, err := listener.Accept(context.Background())

		if err != nil {
			return err
		}

		go server.handleConnection(conn)
	}
}

func (server *WTServer) handleConnection(conn quic.Connection) {

	wtsession, req, err := NewWTSession(conn, conn.Context())

	if err != nil {
		log.Printf("[Error Creating Session][%s]", err)
		return
	}

	log.Println("[New WTSession Created]")

	go func() {
		for {
			buf := make([]byte, 1024)
			_, err := wtsession.rrs.Read(buf)
			if err != nil {
				break
			}
		}
	}()

	server.ServeHTTP(wtsession.ResWriter, req)
}

// Settings({WebTransportMaxSessions: 1, QPackMaxTableCapacity: 0, EnableConnectProtocol: 1, EnableWebTransport: 1, QPackBlockedStreams: 0, H3Datagram: 1})

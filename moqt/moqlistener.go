package moqt

import (
	"context"
	"crypto/tls"

	"moq-go/moqt/wire"

	"moq-go/wt"
	"net/http"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
)

type ListenerOptions struct {
	ListenAddr string
	CertPath   string
	KeyPath    string
	ALPNs      []string
	QuicConfig *quic.Config
}

type MOQTListener struct {
	http.Handler
	Options ListenerOptions
	Ctx     context.Context
	Role    uint64
}

func (listener *MOQTListener) Listen() error {

	Options := listener.Options
	tlsCerts, err := tls.LoadX509KeyPair(Options.CertPath, Options.KeyPath)

	if err != nil {
		return err
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{tlsCerts},
		NextProtos:   Options.ALPNs,
	}

	quiclistener, err := quic.ListenAddr(Options.ListenAddr, &tlsConfig, Options.QuicConfig)

	if err != nil {
		return err
	}

	// WebTransport Handler

	webTransportHandler := func(rw http.ResponseWriter, req *http.Request) {
		wts := req.Body.(*wt.WTSession)
		wts.AcceptSession()

		session, err := CreateMOQSession(wts, listener.Role, SERVER_MODE)

		if err != nil {
			log.Error().Msgf("[Error Creating MOQ Session][%s]", err)
			return
		}

		go session.ServeMOQ()
	}

	listener.Handler = http.HandlerFunc(webTransportHandler)
	http.Handle("/", listener.Handler)
	http.Handle("/moqt", listener.Handler)

	// Now we do the actual listening..

	log.Info().Msgf("[QUIC Listener][Listening on - %s]", Options.ListenAddr)

	for {
		quicConn, err := quiclistener.Accept(listener.Ctx)

		if err != nil {
			log.Debug().Msgf("[QUIC Listener][Error Acceping Quic Conn][%s]", err)
			break
		}

		alpn := quicConn.ConnectionState().TLS.NegotiatedProtocol

		switch alpn {
		case "moq-00":
			listener.handleMOQ(quicConn)
		case "h3":
			listener.handleWebTransport(quicConn)
		default:
			log.Error().Msgf("[QUIC Listener][Uknonwn ALPN - %s]", alpn)
			quicConn.CloseWithError(quic.ApplicationErrorCode(wire.MOQERR_INTERNAL_ERROR), "Internal Error - Unknown ALPN")
		}
	}

	return nil
}

// Handles Plain QUIC based MOQ Sessions
func (listener MOQTListener) handleMOQ(conn quic.Connection) {

	log.Debug().Msgf("[Incoming QUIC Session][IP - %s]", conn.RemoteAddr())

	session, err := CreateMOQSession(conn, listener.Role, SERVER_MODE)

	if err != nil {
		log.Error().Msgf("[Error Creating MOQ Session][%s]", err)
		return
	}

	go session.ServeMOQ()
}

// Handles WebTransport based MOQ Sessions
func (listener MOQTListener) handleWebTransport(conn quic.Connection) {

	log.Info().Msgf("[Incoming WebTransport Session][IP - %s]", conn.RemoteAddr())

	webtransportSession, req, err := wt.UpgradeWTS(conn)

	if err != nil {
		log.Error().Msgf("[Error Upgrading to WTS][%s]", err)
		return
	}

	go listener.ServeHTTP(webtransportSession.ResponseWriter, req)
}

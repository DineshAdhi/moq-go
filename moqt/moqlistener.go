package moqt

import (
	"context"
	"crypto/tls"
	"moq-go/logger"
	"moq-go/wt"
	"net/http"

	"github.com/quic-go/quic-go"
)

type MOQTListener struct {
	http.Handler
	ListenAddr string
	CertPath   string
	KeyPath    string
	ALPNS      []string
	QuicConfig *quic.Config
	Ctx        context.Context
}

func (listener *MOQTListener) Listen() error {

	tlsCerts, err := tls.LoadX509KeyPair(listener.CertPath, listener.KeyPath)

	if err != nil {
		return err
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{tlsCerts},
		NextProtos:   listener.ALPNS,
	}

	quiclistener, err := quic.ListenAddr(listener.ListenAddr, &tlsConfig, listener.QuicConfig)

	if err != nil {
		return err
	}

	// WebTransport Handler

	webTransportHandler := func(rw http.ResponseWriter, req *http.Request) {
		wts := req.Body.(*wt.WTSession)
		wts.AcceptSession()

		moqtsession := CreateMOQSession(wts)
		moqtsession.Serve()
	}

	listener.Handler = http.HandlerFunc(webTransportHandler)
	http.Handle("/", listener.Handler)
	http.Handle("/moqt", listener.Handler)

	// Now we do the actual listening..

	logger.InfoLog("[QUIC Listener][Listening on - %s]", listener.ListenAddr)

	for {
		quicConn, err := quiclistener.Accept(listener.Ctx)

		if err != nil {
			logger.DebugLog("[QUIC Listener][Error Acceping Quic Conn][%s]", err)
			break
		}

		alpn := quicConn.ConnectionState().TLS.NegotiatedProtocol

		switch alpn {
		case "moq-00":
			listener.handleMOQ(quicConn)
		case "h3":
			listener.handleWebTransport(quicConn)
		default:
			logger.ErrorLog("[QUIC Listener][Uknonwn ALPN - %s]", alpn)
			quicConn.CloseWithError(quic.ApplicationErrorCode(MOQERR_INTERNAL_ERROR), "Internal Error - Unknown ALPN")
		}
	}

	return nil
}

// Handles Plain QUIC based MOQ Sessions
func (listener MOQTListener) handleMOQ(conn quic.Connection) {

	logger.DebugLog("[Incoming QUIC Session][IP - %s]", conn.RemoteAddr())

	session := CreateMOQSession(conn)

	go session.Serve()
}

// Handles WebTransport based MOQ Sessions
func (listener MOQTListener) handleWebTransport(conn quic.Connection) {

	logger.DebugLog("[Incoming WebTransport Session][IP - %s]", conn.RemoteAddr())

	webtransportSession, req, err := wt.UpgradeWTS(conn)

	if err != nil {
		logger.ErrorLog("[Error Upgrading to WTS][%s]", err)
		return
	}

	go listener.ServeHTTP(webtransportSession.ResponseWriter, req)
}

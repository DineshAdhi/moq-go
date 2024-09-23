package moqt

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/quic-go/quic-go"
)

type DialerOptions struct {
	KeyPath    string
	CertPath   string
	QuicConfig *quic.Config
	ALPNs      []string
}

type MOQTDialer struct {
	Options DialerOptions
	Role    uint64
	Ctx     context.Context
}

func (d *MOQTDialer) Dial(addr string) (*MOQTSession, error) {

	Options := d.Options

	tlsCerts, err := tls.LoadX509KeyPair(Options.CertPath, Options.KeyPath)

	if err != nil {
		return nil, err
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{tlsCerts},
		NextProtos:   Options.ALPNs,
	}

	conn, err := quic.DialAddr(d.Ctx, addr, &tlsConfig, Options.QuicConfig)

	if err != nil {
		return nil, err
	}

	session, err := CreateMOQSession(conn, d.Role, CLIENT_MODE)

	if err != nil {
		return nil, err
	}

	session.ServeMOQ()

	timeout := time.After(time.Second * 5)

	select {
	case <-session.HandshakeDone:
		return session, nil
	case <-timeout:
		return nil, fmt.Errorf("[Error Dialing MOQT][Timeout]")
	}
}

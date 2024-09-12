package moqt

import (
	"context"
	"crypto/tls"

	"github.com/quic-go/quic-go"
)

type DialerOptions struct {
	DialAddress string
	KeyPath     string
	CertPath    string
	QuicConfig  *quic.Config
	ALPNs       []string
}

type MOQTDialer struct {
	Options DialerOptions
	Role    uint64
	Ctx     context.Context
}

func (d *MOQTDialer) Connect() (*MOQTSession, error) {

	Options := d.Options

	tlsCerts, err := tls.LoadX509KeyPair(Options.CertPath, Options.KeyPath)

	if err != nil {
		return nil, err
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{tlsCerts},
		NextProtos:   Options.ALPNs,
	}

	conn, err := quic.DialAddr(d.Ctx, Options.DialAddress, &tlsConfig, Options.QuicConfig)

	if err != nil {
		return nil, err
	}

	session, err := CreateMOQSession(conn, d.Role, CLIENT_MODE)

	if err != nil {
		return nil, err
	}

	return session, nil
}

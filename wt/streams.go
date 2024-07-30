package wt

import (
	"bytes"
	"fmt"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	STREAM_CONTROL                 = 0x00
	STREAM_PUSH                    = 0x01
	STREAM_QPACK_ENCODER           = 0x02
	STREAM_QPACK_DECODER           = 0x03
	STREAM_WEBTRANSPORT_UNI_STREAM = 0x54
)

type StreamHeader struct {
	Type uint64
	ID   uint64
}

func (s StreamHeader) Read(r io.Reader) error {
	qr := quicvarint.NewReader(r)
	t, err := quicvarint.Read(qr)

	if err != nil {
		return err
	}

	s.Type = t

	switch t {
	// One-byte streams
	case STREAM_CONTROL, STREAM_QPACK_ENCODER, STREAM_QPACK_DECODER:
		return nil
	// Two-byte streams
	case STREAM_PUSH, STREAM_WEBTRANSPORT_UNI_STREAM:
		l, err := quicvarint.Read(qr)
		if err != nil {
			return err
		}
		s.ID = l
		return nil
	default:
		// skip over unknown streams
		return fmt.Errorf("unknown stream type %x", t)
	}
}

func (s StreamHeader) GetBytes() ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.Write(quicvarint.Append(nil, s.Type))
	switch s.Type {
	// One-byte streams
	case STREAM_CONTROL, STREAM_QPACK_ENCODER, STREAM_QPACK_DECODER:
	// Two-byte streams
	case STREAM_PUSH, STREAM_WEBTRANSPORT_UNI_STREAM:
		buf.Write(quicvarint.Append(nil, s.ID))
	default:
		// skip over unknown streams
		return []byte{}, fmt.Errorf("unknown stream type")
	}

	return buf.Bytes(), nil
}

package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	STREAM_CONTROL                 = uint64(0x00)
	STREAM_PUSH                    = uint64(0x01)
	STREAM_QPACK_ENCODER           = uint64(0x02)
	STREAM_QPACK_DECODER           = uint64(0x03)
	STREAM_WEBTRANSPORT_UNI_STREAM = uint64(0x54)
	STREAM_WEBTRANSPORT_BI_STREAM  = uint64(0x41)
)

func GetStreamHeaderString(header uint64) string {
	switch header {
	case STREAM_CONTROL:
		return "STREAM_CONTROL"
	case STREAM_PUSH:
		return "STREAM_PUSH"
	case STREAM_QPACK_ENCODER:
		return "STREAM_QPACK_ENCODER"
	case STREAM_QPACK_DECODER:
		return "STREAM_QPACK_DECODER"
	case STREAM_WEBTRANSPORT_UNI_STREAM:
		return "STREAM_WEBTRANSPORT_UNI_STREAM"
	case STREAM_WEBTRANSPORT_BI_STREAM:
		return "STREAM_WEBTRANSPORT_BI_STREAM"
	default:
		return "UNKNOWN STREAM HEADER"
	}
}

type StreamHeader struct {
	Type uint64
	ID   uint64
}

func (sh *StreamHeader) Read(reader quicvarint.Reader) error {

	htype, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	sh.Type = htype

	switch htype {
	case STREAM_CONTROL, STREAM_QPACK_ENCODER, STREAM_QPACK_DECODER:
		return nil
	case STREAM_PUSH, STREAM_WEBTRANSPORT_UNI_STREAM, STREAM_WEBTRANSPORT_BI_STREAM:
		id, err := quicvarint.Read(reader)
		if err != nil {
			return err
		}
		sh.ID = id
		return nil
	default:
		return fmt.Errorf("[Error Parsing Stream Header][Unknown Type - %X]", htype)
	}
}

func (sh StreamHeader) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, sh.Type)

	switch sh.Type {
	case STREAM_CONTROL, STREAM_QPACK_ENCODER, STREAM_QPACK_DECODER:
	case STREAM_PUSH, STREAM_WEBTRANSPORT_UNI_STREAM:
		data = quicvarint.Append(data, sh.ID)
	}

	return data
}

func (sh StreamHeader) String() string {
	return fmt.Sprintf("[Stream Header][Type - %s][ID - %X]", GetStreamHeaderString(sh.Type), sh.ID)
}

package wire

import (
	"fmt"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

type MOQTStream interface {
	Parse(uint64, quicvarint.Reader) error
	ReadObject() (uint64, *Object, error)
	WriteObject(*Object)
	GetHeaderBytes() []byte
	SetStreamID(string)
	GetStreamID() string
	GetHeaderSubIDBytes(uint64) []byte
	Pipe(int, quic.SendStream) (int, error)
	SetReader(reader quicvarint.Reader)
	WgAdd()
	WgWait()
	WgDone()
	Close()
}

func ParseMOQTStream(reader quicvarint.Reader) (uint64, MOQTStream, error) {

	var stream MOQTStream

	htype, err := quicvarint.Read(reader)

	if err != nil {
		return 0, stream, err
	}

	switch htype {
	case STREAM_HEADER_GROUP, STREAM_HEADER_TRACK:
		break
	default:
		return 0, stream, fmt.Errorf("Unknown Header Type - %X", htype)
	}

	subid, err := quicvarint.Read(reader)

	if err != nil {
		return 0, stream, err
	}

	switch htype {
	case STREAM_HEADER_GROUP:
		stream = &GroupStream{}
	case STREAM_HEADER_TRACK:
		stream = &TrackStream{}
	}

	err = stream.Parse(subid, reader)

	return subid, stream, err
}

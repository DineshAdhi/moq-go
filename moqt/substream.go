package moqt

import (
	"moq-go/moqt/wire"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubStream struct {
	StreamID    string
	SubID       uint64
	StreamsChan chan wire.MOQTStream
}

func NewSubStream(streamid string, subid uint64) *SubStream {
	return &SubStream{
		StreamID:    streamid,
		SubID:       subid,
		StreamsChan: make(chan wire.MOQTStream),
	}
}

func (sub SubStream) GetStreamID() string {
	return sub.StreamID
}

func (sub SubStream) GetSubID() uint64 {
	return sub.SubID
}

func (sub *SubStream) AcceptStream(stream wire.MOQTStream) {
	stream.WgDone()
}

func (sub *SubStream) ProcessObjects(stream wire.MOQTStream, reader quicvarint.Reader) {
	sub.StreamsChan <- stream
}

package wire

import "github.com/quic-go/quic-go"

type MOQTStream interface {
	ReadObject() (uint64, *Object, error)
	GetHeaderBytes() []byte
	GetStreamID() string
	GetHeaderSubIDBytes(subid uint64) []byte
	Pipe(int, quic.SendStream) (int, error)
}

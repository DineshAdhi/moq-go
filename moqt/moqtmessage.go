package moqt

import "io"

type MOQTMessage interface {
	Parse(r io.Reader) error
	GetBytes() []byte
}

type MOQTReader struct {
	io.Reader
	io.ByteReader
}

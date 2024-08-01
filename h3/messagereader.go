package h3

import (
	"io"
)

type MessageReader interface {
	io.Reader
	io.ByteReader
}

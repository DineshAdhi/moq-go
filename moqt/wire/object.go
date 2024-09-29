package wire

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type Object struct {
	ID      uint64
	Payload []byte
}

func (object *Object) Parse(reader quicvarint.Reader) error {

	var err error

	if object.ID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	length, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	object.Payload = make([]byte, length)
	io.ReadFull(reader, object.Payload)

	return nil
}

func (object *Object) GetBytes() []byte {
	var data []byte

	length := len(object.Payload)

	data = quicvarint.Append(data, object.ID)
	data = quicvarint.Append(data, uint64(length))
	data = append(data, object.Payload...)

	return data
}

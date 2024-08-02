package h3

import (
	"github.com/quic-go/quic-go/quicvarint"
)

type DataFrame struct {
	data []byte
}

func (dframe *DataFrame) Parse(reader quicvarint.Reader) error {

	len, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	data := make([]byte, len)
	reader.Read(data)

	dframe.data = data

	return nil
}

func (dframe DataFrame) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, FRAME_DATA)
	data = quicvarint.Append(data, uint64(len(dframe.data)))
	data = append(data, dframe.data...)

	return data
}

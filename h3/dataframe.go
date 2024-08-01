package h3

import "github.com/quic-go/quic-go/quicvarint"

type DataFrame struct {
	data []byte
}

func (dframe *DataFrame) Parse(r MessageReader) error {
	len, err := quicvarint.Read(r)

	if err != nil {
		return err
	}

	data := make([]byte, len)
	r.Read(data)

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

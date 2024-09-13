package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubscribeDone struct {
	SubscribeID   uint64
	StatusCode    uint64
	ReasonPhrase  string
	ContentExists uint8
	FinalGroup    uint64
	FinalObject   uint64
}

func (m *SubscribeDone) Parse(reader quicvarint.Reader) (err error) {

	if m.SubscribeID, err = quicvarint.Read(reader); err != nil {
		return
	}

	if m.StatusCode, err = quicvarint.Read(reader); err != nil {
		return
	}

	if m.ReasonPhrase, err = ParseVarIntString(reader); err != nil {
		return
	}

	if m.ContentExists, err = reader.ReadByte(); err != nil {
		return
	}

	if m.ContentExists == 1 {

		if m.FinalGroup, err = quicvarint.Read(reader); err != nil {
			return
		}

		if m.FinalObject, err = quicvarint.Read(reader); err != nil {
			return
		}
	}

	return nil
}

func (m *SubscribeDone) GetBytes() []byte {
	var data []byte

	reason := []byte(m.ReasonPhrase)

	data = quicvarint.Append(data, m.SubscribeID)
	data = quicvarint.Append(data, m.StatusCode)
	data = quicvarint.Append(data, uint64(len(m.ReasonPhrase)))
	data = append(data, reason...)
	data = append(data, m.ContentExists)

	if m.ContentExists == 1 {
		data = quicvarint.Append(data, m.FinalGroup)
		data = quicvarint.Append(data, m.FinalObject)
	}

	return data
}

func (m *SubscribeDone) String() string {

	str := fmt.Sprintf("[SUBSCRIBE DONE][ID - %X][Code - %d][%s]", m.SubscribeID, m.StatusCode, m.ReasonPhrase)

	if m.ContentExists == 1 {
		str += fmt.Sprintf("[Final Group - %d][Final Object - %d]", m.FinalGroup, m.FinalObject)
	}

	return str
}

func (m *SubscribeDone) Type() uint64 {
	return SUBSCRIBE_DONE
}

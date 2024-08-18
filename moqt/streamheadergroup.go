package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type StreamHeaderGroupMessage struct {
	SubscribeID uint64
	TrackAlias  uint64
	GroupID     uint64
	SendOrder   uint64
}

func (shg *StreamHeaderGroupMessage) Type() uint64 {
	return STREAM_HEADER_GROUP
}

func (shg *StreamHeaderGroupMessage) Parse(reader quicvarint.Reader) (err error) {

	if shg.SubscribeID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if shg.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if shg.GroupID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if shg.SendOrder, err = quicvarint.Read(reader); err != nil {
		return err
	}

	return nil
}

func (shg *StreamHeaderGroupMessage) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, STREAM_HEADER_GROUP)
	data = quicvarint.Append(data, shg.SubscribeID)
	data = quicvarint.Append(data, shg.TrackAlias)
	data = quicvarint.Append(data, shg.GroupID)
	data = quicvarint.Append(data, shg.SendOrder)

	return data
}

func (shg *StreamHeaderGroupMessage) String() string {
	return fmt.Sprintf("[%s][ID - %X][Group ID - %X][Track Alias - %X][Send Order - %X]", GetMoqMessageString(shg.Type()), shg.SubscribeID, shg.GroupID, shg.TrackAlias, shg.SendOrder)
}

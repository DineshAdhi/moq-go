package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

// SUBSCRIBE_OK
// {
//   Subscribe ID (i),
//   Expires (i),
//   ContentExists (1),
//   [Largest Group ID (i)],
//   [Largest Object ID (i)]
// }

type SubsribeOkMessage struct {
	id              uint64
	expires         uint64
	contentexists   uint8
	largestGroupId  uint64
	largestObjectId uint64
}

func (s SubsribeOkMessage) Type() uint64 {
	return SUBSCRIBE_OK
}

func (s *SubsribeOkMessage) Parse(reader quicvarint.Reader) (err error) {

	if s.id, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.expires, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.contentexists, err = reader.ReadByte(); err != nil {
		return err
	}

	if s.contentexists == 1 {
		if s.largestObjectId, err = quicvarint.Read(reader); err != nil {
			return err
		}

		if s.largestObjectId, err = quicvarint.Read(reader); err != nil {
			return err
		}
	}

	return nil
}

func (s SubsribeOkMessage) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, SUBSCRIBE_OK)
	data = quicvarint.Append(data, s.id)
	data = quicvarint.Append(data, s.expires)
	data = append(data, s.contentexists)

	if s.contentexists == 1 {
		data = quicvarint.Append(data, s.largestObjectId)
		data = quicvarint.Append(data, s.largestGroupId)
	}

	return data
}

func (s SubsribeOkMessage) String() string {
	return fmt.Sprintf("[SubscribeOK][ID - %X][Expires - %d][Content Exists - %d][L Group Id / Object Id - %d / %d]", s.id, s.expires, s.contentexists, s.largestObjectId, s.largestGroupId)
}

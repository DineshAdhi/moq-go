package wire

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

type SubscribeOk struct {
	SubscribeID     uint64
	Expires         uint64
	ContentExists   uint8
	largestGroupId  uint64
	largestObjectId uint64
}

func (s SubscribeOk) Type() uint64 {
	return SUBSCRIBE_OK
}

func (s *SubscribeOk) Parse(reader quicvarint.Reader) (err error) {

	if s.SubscribeID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.Expires, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.ContentExists, err = reader.ReadByte(); err != nil {
		return err
	}

	if s.ContentExists == 1 {
		if s.largestObjectId, err = quicvarint.Read(reader); err != nil {
			return err
		}

		if s.largestObjectId, err = quicvarint.Read(reader); err != nil {
			return err
		}
	}

	return nil
}

func (s SubscribeOk) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, SUBSCRIBE_OK)
	data = quicvarint.Append(data, s.SubscribeID)
	data = quicvarint.Append(data, s.Expires)
	data = append(data, s.ContentExists)

	if s.ContentExists == 1 {
		data = quicvarint.Append(data, s.largestObjectId)
		data = quicvarint.Append(data, s.largestGroupId)
	}

	return data
}

func (s SubscribeOk) String() string {
	return fmt.Sprintf("[SubscribeOK][ID - %X][Expires - %d][Content Exists - %d][L Group Id / Object Id - %d / %d]", s.SubscribeID, s.Expires, s.ContentExists, s.largestObjectId, s.largestGroupId)
}

func GetSubOKMessage(id uint64) SubscribeOk {
	okmsg := SubscribeOk{}
	okmsg.SubscribeID = id
	okmsg.Expires = 0
	okmsg.ContentExists = 1
	okmsg.largestGroupId = 0
	okmsg.largestObjectId = 0

	return okmsg
}

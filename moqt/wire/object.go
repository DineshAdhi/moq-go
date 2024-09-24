package wire

import "github.com/quic-go/quic-go/quicvarint"

// {
//   Object ID (i),
//   Object Payload Length (i),
//   [Object Status (i)],
//   Object Payload (..),
// }

type TrackObject struct {
	ObjectID uint64
	Length   uint64
	Status   uint64
	Payload  []byte
}

func (obj *TrackObject) Parse(reader quicvarint.Reader) error {

	var err error

	if obj.ObjectID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if obj.Length, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if obj.Length == 0 {
		if obj.Status, err = quicvarint.Read(reader); err != nil {
			return err
		}
	} else {
		obj.Payload = make([]byte, obj.Length)
		reader.Read(obj.Payload)
	}

	return nil
}

func (obj *TrackObject) GetBytes() []byte {

	obj.Length = uint64(len(obj.Payload))

	var payload []byte
	payload = quicvarint.Append(payload, obj.ObjectID)
	payload = quicvarint.Append(payload, obj.Length)

	if obj.Length == 0 {
		payload = quicvarint.Append(payload, obj.Status)
	}

	payload = append(payload, obj.Payload...)

	return payload
}

package wire

import (
	"io"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	OBJECT_EXPIRY_TIME = 10
)

type TrackObject struct {
	ObjectID  uint64
	Length    uint64
	Status    uint64
	Payload   []byte
	CreatedAt time.Time
	StreamID  string
	Header    MOQTObjectHeader
}

func NewTrackObject(streamid string, header MOQTObjectHeader) *TrackObject {
	object := &TrackObject{}
	object.CreatedAt = time.Now()
	object.StreamID = streamid
	object.Header = header

	return object
}

func (object *TrackObject) GetStreamID() string {
	return object.StreamID
}

func (object *TrackObject) IsExpired() bool {
	now := time.Now()

	if now.Sub(object.CreatedAt).Seconds() >= OBJECT_EXPIRY_TIME {
		return true
	}

	return false
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
		_, err := io.ReadFull(reader, obj.Payload)

		if err != nil {
			return err
		}
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

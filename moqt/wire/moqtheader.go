package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	OBJECT_STREAM       = 0x0
	OBJECT_DATAGRAM     = 0x1
	STREAM_HEADER_TRACK = 0x50
	STREAM_HEADER_GROUP = 0x51
)

type MOQTObjectHeader interface {
	Type() uint64
	GetSubID() uint64
	GetTrackAlias() uint64
	GetGroupKey() string
	GetGroupID() uint64
	Parse(quicvarint.Reader) error
	GetBytes(uint64) []byte
	String() string
}

func ParseMOQTObjectHeader(reader quicvarint.Reader) (MOQTObjectHeader, error) {
	var objectHeader MOQTObjectHeader

	htype, err := quicvarint.Read(reader)

	if err != nil {
		return nil, fmt.Errorf("[Error Reading Header Type][%s]", err)
	}

	switch htype {
	case STREAM_HEADER_GROUP:
		objectHeader = &GroupHeader{}
	default:
		return nil, fmt.Errorf("[Unknown Object Header Type][%X]", htype)
	}

	err = objectHeader.Parse(reader)

	if err != nil {
		return nil, fmt.Errorf("[Error While Parsing Stream Header][Type - %X][%s]", htype, err)
	}

	// log.Debug().Msgf("[MOQT ObjStream Header][%s]", objectHeader)

	return objectHeader, nil
}

type GroupHeader struct {
	SubscribeID uint64
	TrackAlias  uint64
	GroupID     uint64
	SendOrder   uint64
}

func (gh GroupHeader) Type() uint64 {
	return STREAM_HEADER_GROUP
}

func (gh *GroupHeader) GetSubID() uint64 {
	return gh.SubscribeID
}

func (gh *GroupHeader) GetGroupID() uint64 {
	return gh.GroupID
}

func (gh *GroupHeader) GetTrackAlias() uint64 {
	return gh.TrackAlias
}

func (gh *GroupHeader) GetGroupKey() string {
	return fmt.Sprintf("%X_%d", gh.TrackAlias, gh.GroupID)
}

func (gh *GroupHeader) Parse(reader quicvarint.Reader) (err error) {

	if gh.SubscribeID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if gh.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if gh.GroupID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if gh.SendOrder, err = quicvarint.Read(reader); err != nil {
		return err
	}

	return nil
}

func (gh *GroupHeader) GetBytes(id uint64) []byte {
	var data []byte
	data = quicvarint.Append(data, STREAM_HEADER_GROUP)
	data = quicvarint.Append(data, id)
	data = quicvarint.Append(data, gh.TrackAlias)
	data = quicvarint.Append(data, gh.GroupID)
	data = quicvarint.Append(data, gh.SendOrder)

	return data
}

func (gh *GroupHeader) String() string {
	return fmt.Sprintf("[%s][ID - %X][Group ID - %X][Track Alias - %X][Send Order - %X]", GetMoqMessageString(gh.Type()), gh.SubscribeID, gh.GroupID, gh.TrackAlias, gh.SendOrder)
}

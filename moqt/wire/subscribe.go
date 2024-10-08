package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	LatestGroup   = uint64(0x1)
	LatestObject  = uint64(0x2)
	AbsoluteStart = uint64(0x3)
	AbsoluteRange = uint64(0x4)
)

type StreamID struct {
	string
}

func GetFilterType(ftype uint64) string {
	switch ftype {
	case LatestGroup:
		return "LatestGroup"
	case LatestObject:
		return "LatestObject"
	case AbsoluteStart:
		return "AbsoluteStart"
	case AbsoluteRange:
		return "AbsoluteRange"
	default:
		return "UKNOWN FILTER TYPE"
	}
}

type Subscribe struct {
	SubscribeID    uint64
	TrackAlias     uint64
	TrackNameSpace string
	TrackName      string
	FilterType     uint64
	StartGroup     uint64
	StartObject    uint64
	EndGroup       uint64
	EndObject      uint64
	Params         Parameters
}

func (s *Subscribe) Parse(reader quicvarint.Reader) (err error) {

	if s.SubscribeID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.TrackNameSpace, err = ParseVarIntString(reader); err != nil {
		return err
	}

	if s.TrackName, err = ParseVarIntString(reader); err != nil {
		return err
	}

	if s.FilterType, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.FilterType == AbsoluteStart || s.FilterType == AbsoluteRange {

		s.StartGroup, err = quicvarint.Read(reader)

		if err != nil {
			return err
		}
	}

	if s.FilterType == AbsoluteStart || s.FilterType == AbsoluteRange {
		s.StartObject, err = quicvarint.Read(reader)

		if err != nil {
			return err
		}
	}

	if s.FilterType == AbsoluteRange {
		s.EndGroup, err = quicvarint.Read(reader)

		if err != nil {
			return err
		}
	}

	if s.FilterType == AbsoluteRange {
		s.EndObject, err = quicvarint.Read(reader)

		if err != nil {
			return err
		}
	}

	params := Parameters{}
	err = params.Parse(reader)

	if err != nil {
		return err
	}

	s.Params = params

	return nil
}

func (s Subscribe) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, SUBSCRIBE)
	data = quicvarint.Append(data, s.SubscribeID)
	data = quicvarint.Append(data, s.TrackAlias)
	data = append(data, GetBytesVarIntString(s.TrackNameSpace)...)
	data = append(data, GetBytesVarIntString(s.TrackName)...)
	data = quicvarint.Append(data, s.FilterType)

	if s.FilterType == AbsoluteStart || s.FilterType == AbsoluteRange {
		data = quicvarint.Append(data, s.StartGroup)
	}

	if s.FilterType == AbsoluteStart || s.FilterType == AbsoluteRange {
		data = quicvarint.Append(data, s.StartObject)
	}

	if s.FilterType == AbsoluteRange {
		data = quicvarint.Append(data, s.EndGroup)
	}

	if s.FilterType == AbsoluteRange {
		data = quicvarint.Append(data, s.EndObject)
	}

	data = append(data, s.Params.GetBytes()...)

	return data
}

// Stream ID is a concat of namespace + ObjectStream + alias. It makes it unique across all sessions
func (s Subscribe) GetStreamID() string {
	return fmt.Sprintf("%s_%s", s.TrackNameSpace, s.TrackName)
}

func (s Subscribe) String() string {
	str := fmt.Sprintf("[%s][ID - %X][Filter Type - %s][Name - %s][Alias - %X][NameSpace - %s]", GetMoqMessageString(SUBSCRIBE), s.SubscribeID, GetFilterType(s.FilterType), s.TrackName, s.TrackAlias, s.TrackNameSpace)

	if len(s.Params) > 0 {
		str += s.Params.String()
	}

	return str
}

func (s Subscribe) Type() uint64 {
	return SUBSCRIBE
}

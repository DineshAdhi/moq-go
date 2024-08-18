package moqt

import (
	"fmt"
	"strconv"

	"github.com/quic-go/quic-go/quicvarint"
)

// SUBSCRIBE Message {
// 	Subscribe ID (i),
// 	Track Alias (i),
// 	Track Namespace (b),
// 	Track Name (b),
// 	Filter Type (i),
// 	[StartGroup (i),
// 	 StartObject (i)],
// 	[EndGroup (i),
// 	 EndObject (i)],
// 	Number of Parameters (i),
// 	Subscribe Parameters (..) ...
//   }

const (
	LatestGroup   = uint64(0x1)
	LatestObject  = uint64(0x2)
	AbsoluteStart = uint64(0x3)
	AbsoluteRange = uint64(0x4)
)

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

type SubscribeMessage struct {
	SubscribeID    uint64
	TrackAlias     uint64
	TrackNamespace string
	TrackName      string
	FilterType     uint64
	StartGroup     uint64
	StartObject    uint64
	EndGroup       uint64
	EndObject      uint64
	Params         Parameters
}

func (s *SubscribeMessage) Parse(reader quicvarint.Reader) (err error) {

	if s.SubscribeID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.TrackNamespace, err = ParseVarIntString(reader); err != nil {
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

func (s SubscribeMessage) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, SUBSCRIBE)
	data = quicvarint.Append(data, s.SubscribeID)
	data = quicvarint.Append(data, s.TrackAlias)
	data = append(data, GetBytesVarIntString(s.TrackNamespace)...)
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

// Stream ID is a concat of namespace + track + alias. It makes it unique across all sessions
func (s SubscribeMessage) getCacheKey() string {
	return fmt.Sprintf("%s_%s_%s", s.TrackNamespace, s.TrackName, strconv.Itoa(int(s.TrackAlias)))
}

func (s SubscribeMessage) String() string {
	str := fmt.Sprintf("[%s][ID - %X][Filter Type - %s][Track Name - %s][Track Alias - %X][Name Space - %s]", GetMoqMessageString(SUBSCRIBE), s.SubscribeID, GetFilterType(s.FilterType), s.TrackName, s.TrackAlias, s.TrackNamespace)

	if len(s.Params) > 0 {
		str += s.Params.String()
	}

	return str
}

func (s SubscribeMessage) Type() uint64 {
	return SUBSCRIBE
}

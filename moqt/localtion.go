package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type Location struct {
	Mode  uint64
	Value uint64
}

const (
	LOCATION_NONE        = uint64(0x0)
	LOCATION_ABSOLUTE    = uint64(0x1)
	LOCATION_RELPREVIOUS = uint64(0x2)
	LOCATION_RELPRENEXT  = uint64(0x3)
)

func GetLocatonString(mode uint64) string {
	switch mode {
	case LOCATION_NONE:
		return "LOCATION_NONE"
	case LOCATION_ABSOLUTE:
		return "LOCATION_ABSOLUTE"
	case LOCATION_RELPREVIOUS:
		return "LOCATION_RELPREVIOUS"
	case LOCATION_RELPRENEXT:
		return "LOCATION_RELPRENEXT"
	default:
		return "Unknown Location"
	}
}

func (l *Location) Parse(reader quicvarint.Reader) (err error) {

	if l.Mode, err = quicvarint.Read(reader); err != nil {
		return nil
	}

	switch l.Mode {
	case LOCATION_NONE:
		return nil
	case LOCATION_ABSOLUTE, LOCATION_RELPREVIOUS, LOCATION_RELPRENEXT:
		if l.Value, err = quicvarint.Read(reader); err != nil {
			return err
		}
	}

	return nil
}

func (l Location) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, l.Mode)

	switch l.Mode {
	case LOCATION_NONE:
		return data
	case LOCATION_ABSOLUTE, LOCATION_RELPREVIOUS, LOCATION_RELPRENEXT:
		data = quicvarint.Append(data, l.Value)
	}

	return data
}

func (l Location) String() string {
	return fmt.Sprintf("%s : %x", GetLocatonString(l.Mode), l.Value)
}

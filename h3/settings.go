package h3

import (
	"bytes"
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	SETTINGS_MAX_FIELD_SECTION_SIZE   = uint64(0x6)
	SETTINGS_QPACK_MAX_TABLE_CAPACITY = uint64(0x1)
	H3_DATAGRAM_05                    = uint64(0xffd277)
	ENABLE_WEBTRANSPORT               = uint64(0x2b603742)
	WEBTRANSPORT_MAX_SESSIONS         = uint64(0xc671706a)
	SETTINGS_H3_DATAGRAM              = uint64(0x33)
	SETTINGS_ENABLE_CONNECT_PROTOCOL  = uint64(0x08)
	SETTINGS_QPACK_BLOCKED_STREAMS    = uint64(0x07)
)

func GetSettingString(stype uint64) string {
	switch stype {
	case SETTINGS_MAX_FIELD_SECTION_SIZE:
		return "SETTINGS_MAX_FIELD_SECTION_SIZE"
	case SETTINGS_QPACK_MAX_TABLE_CAPACITY:
		return "SETTINGS_QPACK_MAX_TABLE_CAPACITY"
	case H3_DATAGRAM_05:
		return "H3_DATAGRAM_05"
	case ENABLE_WEBTRANSPORT:
		return "ENABLE_WEBTRANSPORT"
	case WEBTRANSPORT_MAX_SESSIONS:
		return "WEBTRANSPORT_MAX_SESSIONS"
	case SETTINGS_H3_DATAGRAM:
		return "SETTINGS_H3_DATAGRAM"
	case SETTINGS_ENABLE_CONNECT_PROTOCOL:
		return "SETTINGS_ENABLE_CONNECT_PROTOCOL"
	case SETTINGS_QPACK_BLOCKED_STREAMS:
		return "SETTINGS_QPACK_BLOCKED_STREAMS"
	default:
		return "Unknown Setting Type"
	}
}

type Setting struct {
	Key   uint64
	Value uint64
}

type SettingsFrame struct {
	Type     uint64
	Length   uint64
	Settings []Setting
}

func (sframe *SettingsFrame) Read(f *Frame) error {

	reader := bytes.NewReader(f.fpayload)

	sarr := []Setting{}

	for reader.Len() > 0 {
		key, err := quicvarint.Read(reader)

		if err != nil {
			return err
		}

		value, err := quicvarint.Read(reader)

		if err != nil {
			return err
		}

		sarr = append(sarr, Setting{key, value})
	}

	sframe.Type = f.Type
	sframe.Length = f.flength
	sframe.Settings = sarr

	return nil
}

func (sframe *SettingsFrame) GetBytes() []byte {

	var length uint64 = 0

	for _, s := range sframe.Settings {
		length += uint64(quicvarint.Len(s.Key) + quicvarint.Len(s.Value))
	}

	var data []byte

	for _, s := range sframe.Settings {
		data = quicvarint.Append(data, s.Key)
		data = quicvarint.Append(data, s.Value)
	}

	f := Frame{Type: FRAME_SETTINGS, flength: length, fpayload: data}

	return f.GetBytes()
}

func (sframe *SettingsFrame) GetString() string {

	str := fmt.Sprintf("[Type - %s][Length - %d][{", GetFrameTypeString(sframe.Type), sframe.Length)

	for _, s := range sframe.Settings {
		str += fmt.Sprintf("%s : %d ", GetSettingString(s.Key), s.Value)
	}

	str += "}]"

	return str
}

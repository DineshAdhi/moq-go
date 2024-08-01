package h3

import (
	"bytes"
	"fmt"
	"strings"

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

func (setting *Setting) Parse(r MessageReader) error {
	key, err := quicvarint.Read(r)

	if err != nil {
		return err
	}

	value, err := quicvarint.Read(r)

	if err != nil {
		return err
	}

	setting.Key = key
	setting.Value = value

	return nil
}

func (setting Setting) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, setting.Key)
	data = quicvarint.Append(data, setting.Value)

	return data
}

func (setting Setting) GetLength() uint64 {
	var len uint64 = 0
	len = uint64(quicvarint.Len(setting.Key) + quicvarint.Len(setting.Value))
	return len
}

type SettingsFrame struct {
	Settings []Setting
}

func (sframe *SettingsFrame) Parse(r MessageReader) error {
	length, err := quicvarint.Read(r)

	if err != nil {
		return err
	}

	data := make([]byte, length)
	_, err = r.Read(data)

	if err != nil {
		return err
	}

	dlength := uint64(length)
	reader := bytes.NewReader(data)

	for dlength > 0 {
		setting := Setting{}
		setting.Parse(reader)

		sframe.Settings = append(sframe.Settings, setting)

		dlength = dlength - setting.GetLength()
	}

	return nil
}

func (sframe SettingsFrame) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, FRAME_SETTINGS)

	var length uint64 = 0

	for _, setting := range sframe.Settings {
		length += setting.GetLength()
	}

	data = quicvarint.Append(data, length)

	for _, setting := range sframe.Settings {
		data = append(data, setting.GetBytes()...)
	}

	return data
}

func (sframe *SettingsFrame) GetString() string {

	str := "{"

	for _, s := range sframe.Settings {
		str += fmt.Sprintf("%s : %d ", GetSettingString(s.Key), s.Value)
	}

	str += strings.TrimSuffix(str, " ") + "}"

	return str
}

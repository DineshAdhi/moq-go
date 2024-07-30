package wt

import (
	"bytes"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	// https://datatracker.ietf.org/doc/html/draft-ietf-quic-http-34
	SETTINGS_MAX_FIELD_SECTION_SIZE = uint64(0x6)
	// https://datatracker.ietf.org/doc/html/draft-ietf-quic-qpack-21
	SETTINGS_QPACK_MAX_TABLE_CAPACITY = uint64(0x1)
	// https://datatracker.ietf.org/doc/html/draft-ietf-masque-h3-datagram-05#section-9.1
	H3_DATAGRAM_05 = uint64(0xffd277)
	// https://www.ietf.org/archive/id/draft-ietf-webtrans-http3-02.html#section-8.2
	ENABLE_WEBTRANSPORT              = uint64(0x2b603742)
	WEBTRANSPORT_MAX_SESSIONS        = uint64(0xc671706a)
	SETTINGS_H3_DATAGRAM             = uint64(0x33)
	SETTINGS_ENABLE_CONNECT_PROTOCOL = uint64(0x08)
	SETTINGS_QPACK_BLOCKED_STREAMS   = uint64(0x07)
)

// SETTINGS_ENABLE_CONNECT_PROTOCOL,SETTINGS_H3_DATAGRAM,WEBTRANSPORT_MAX_SESSIONS,ENABLE_WEBTRANSPORT,H3_DATAGRAM_05

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

type Settings struct {
	values map[uint64]uint64
}

func DefaultSettings() Settings {

	return Settings{
		values: map[uint64]uint64{
			ENABLE_WEBTRANSPORT:              1,
			SETTINGS_H3_DATAGRAM:             1,
			WEBTRANSPORT_MAX_SESSIONS:        1,
			SETTINGS_ENABLE_CONNECT_PROTOCOL: 1,
			// H3_DATAGRAM_05:                   1,
		},
	}
}

func (s Settings) ToFrame() Frame {

	f := Frame{Type: FRAME_SETTINGS}

	var l uint64 = 0

	for id, value := range s.values {
		l += uint64(quicvarint.Len(uint64(id)) + quicvarint.Len(value))
	}

	f.Length = l

	data := &bytes.Buffer{}
	for id, value := range s.values {
		data.Write(quicvarint.Append(nil, id))
		data.Write(quicvarint.Append(nil, value))
	}

	f.Data = data.Bytes()

	return f
}

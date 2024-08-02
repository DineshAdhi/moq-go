package h3

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	FRAME_DATA                    = uint64(0x00)
	FRAME_HEADERS                 = uint64(0x01)
	FRAME_CANCEL_PUSH             = uint64(0x03)
	FRAME_SETTINGS                = uint64(0x04)
	FRAME_PUSH_PROMISE            = uint64(0x05)
	FRAME_GOAWAY                  = uint64(0x07)
	FRAME_MAX_PUSH_ID             = uint64(0x0D)
	FRAME_WEBTRANSPORT_UNI_STREAM = uint64(0x54)
	FRAME_WEBTRANSPORT_BI_STREAM  = uint64(0x41)
)

func GetFrameTypeString(ftype uint64) string {
	switch ftype {
	case FRAME_DATA:
		return string("FRAME_DATA")
	case FRAME_HEADERS:
		return string("FRAME_HEADERS")
	case FRAME_CANCEL_PUSH:
		return string("FRAME_CANCEL_PUSH")
	case FRAME_SETTINGS:
		return string("FRAME_SETTINGS")
	case FRAME_PUSH_PROMISE:
		return string("FRAME_PUSH_PROMISE")
	case FRAME_GOAWAY:
		return string("FRAME_GOAWAY")
	case FRAME_MAX_PUSH_ID:
		return string("FRAME_MAX_PUSH_ID")
	case FRAME_WEBTRANSPORT_UNI_STREAM:
		return string("FRAME_WEBTRANSPORT_UNI_STREAM")
	case FRAME_WEBTRANSPORT_BI_STREAM:
		return string("FRAME_WEBTRANSPORT_BI_STREAM")
	default:
		return string("Unknown Header Type")
	}
}

type Frame interface {
	Parse(reader quicvarint.Reader) error
	GetBytes() []byte
}

func ParseFrame(reader quicvarint.Reader) (uint64, Frame, error) {

	ftype, err := quicvarint.Read(reader)

	if err != nil {
		return 0, nil, err
	}

	var frame Frame

	switch ftype {
	case FRAME_SETTINGS:
		frame = &SettingsFrame{}
	case FRAME_HEADERS:
		frame = &HeaderFrame{}
	case FRAME_DATA, FRAME_WEBTRANSPORT_BI_STREAM, FRAME_WEBTRANSPORT_UNI_STREAM:
		frame = &DataFrame{}
	default:
		len, err := quicvarint.Read(reader)

		if err != nil {
			return ftype, nil, err
		}

		data := make([]byte, len)
		reader.Read(data)
		return ftype, nil, fmt.Errorf("[Unkown Frame][Type - %X]", ftype)
	}

	err = frame.Parse(reader)

	if err != nil {
		return ftype, nil, err
	}

	return ftype, frame, nil
}

package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	SUBSCRIBE            = 0x3
	SUBSCRIBE_OK         = 0x4
	SUBSCRIBE_ERROR      = 0x5
	ANNOUNCE             = 0x6
	ANNOUNCE_OK          = 0x7
	ANNOUNCE_ERROR       = 0x8
	UNANNOUNCE           = 0x9
	UNSUBSCRIBE          = 0xA
	SUBSCRIBE_DONE       = 0xB
	ANNOUNCE_CANCEL      = 0xC
	TRACK_STATUS_REQUEST = 0xD
	TRACK_STATUS         = 0xE
	GOAWAY               = 0x10
	CLIENT_SETUP         = 0x40
	SERVER_SETUP         = 0x41
)

func GetMoqMessageString(mtype uint64) string {
	switch mtype {
	case SUBSCRIBE:
		return "SUBSCRIBE"
	case SUBSCRIBE_OK:
		return "SUBSCRIBE_OK"
	case SUBSCRIBE_ERROR:
		return "SUBSCRIBE_ERROR"
	case ANNOUNCE:
		return "ANNOUNCE"
	case ANNOUNCE_OK:
		return "ANNOUNCE_OK"
	case ANNOUNCE_ERROR:
		return "ANNOUNCE_ERROR"
	case UNANNOUNCE:
		return "UNANNOUNCE"
	case UNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case SUBSCRIBE_DONE:
		return "SUBSCRIBE_DONE"
	case ANNOUNCE_CANCEL:
		return "ANNOUNCE_CANCEL"
	case TRACK_STATUS_REQUEST:
		return "TRACK_STATUS_REQUEST"
	case TRACK_STATUS:
		return "TRACK_STATUS"
	case GOAWAY:
		return "GOAWAY"
	case CLIENT_SETUP:
		return "CLIENT_SETUP"
	case SERVER_SETUP:
		return "SERVER_SETUP"
	case STREAM_HEADER_TRACK:
		return "STREAM_HEADER_TRACK"
	case STREAM_HEADER_GROUP:
		return "STREAM_HEADER_GROUP"
	default:
		return "UNKNOWN_MESSAGE_TYPE"
	}
}

type MOQTMessage interface {
	Parse(reader quicvarint.Reader) error
	GetBytes() []byte
	String() string
	Type() uint64
}

func ParseMOQTMessage(reader quicvarint.Reader) (MOQTMessage, error) {

	mtype, err := quicvarint.Read(reader)

	if err != nil {
		return nil, err
	}

	var msg MOQTMessage

	switch mtype {
	case CLIENT_SETUP:
		msg = &ClientSetup{}
	case SERVER_SETUP:
		msg = &ServerSetup{}
	case ANNOUNCE:
		msg = &AnnounceMessage{}
	case SUBSCRIBE:
		msg = &SubscribeMessage{}
	case SUBSCRIBE_OK:
		msg = &SubscribeOkMessage{}
	default:
		return nil, fmt.Errorf("unkown MOQT Message %d %+v", mtype, msg)
	}

	msg.Parse(reader)

	return msg, nil
}

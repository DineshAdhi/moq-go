package moqt

import (
	"moq-go/h3"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	ROLE_PARAM = uint64(0x00)
	ROLE_PATH  = uint64(0x01)
)

const (
	ROLE_PUBLISHER  = uint64(0x01)
	ROLE_SUBSCRIBER = uint64(0x02)
	ROLE_PUBSUB     = uint64(0x03)
)

type Parameter interface {
	Parse(r h3.MessageReader) error
	Type() uint64
	Value() interface{}
	GetBytes() []byte
	String() string
}

type Parameters map[uint64]Parameter

func GetParamKeyString(param Parameter) string {
	switch param.Type() {
	case ROLE_PARAM:
		return "ROLE"
	case ROLE_PATH:
		return "PATH"
	default:
		return "UNKNOWN PARAM"
	}
}

func GetRoleString(param Parameter) string {
	switch param.Value().(uint64) {
	case ROLE_PUBLISHER:
		return string("ROLE_PUBLISHER")
	case ROLE_SUBSCRIBER:
		return string("ROLE_PUBLISHER")
	case ROLE_PUBSUB:
		return string("ROLE_PUBLISHER")
	default:
		return string("UNKNOWN_ROLE")
	}
}

func GetParamValueString(param Parameter) string {
	switch param.Type() {
	case ROLE_PARAM:
		return GetRoleString(param)
	default:
		return "UNKNOWN_VALUE"
	}
}

func (params Parameters) Parse(reader h3.MessageReader) error {

	length, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	for range length {
		ptype, err := quicvarint.Read(reader)

		if err != nil {
			return err
		}

		switch ptype {
		case ROLE_PARAM:
			intParam := &IntParameter{ptype: ptype}
			err := intParam.Parse(reader)

			if err != nil {
				return err
			}

			params[ptype] = intParam
		case ROLE_PATH:
			strParam := &StringParameter{ptype: ptype}
			err := strParam.Parse(reader)

			if err != nil {
				return err
			}

			params[ptype] = strParam
		default:
			len, err := quicvarint.Read(reader)
			if err != nil {
				return err
			}

			discardData := make([]byte, len)
			reader.Read(discardData)
		}
	}

	return nil
}

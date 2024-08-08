package moqt

import (
	"fmt"

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

const (
	DRAFT_00 = 0xff000000
	DRAFT_01 = 0xff000001
	DRAFT_02 = 0xff000002
	DRAFT_03 = 0xff000003
)

const (
	MOQERR_NOERROR               = uint64(0x0)
	MOQERR_INTERNAL_ERROR        = uint64(0x1)
	MOQERR_UNAUTHORIZED          = uint64(0x2)
	MOQERR_PROTOCOL_VIOLATION    = uint64(0x3)
	MOQERR_DUPLICATE_TRACK_ALIAS = uint64(0x4)
	MOQERR_PARAM_LENGTH_MISMATCH = uint64(0x5)
	MOQERR_GOAWAY_TIMEOUT        = uint64(0x10)
)

type Parameter interface {
	Parse(reader quicvarint.Reader) error
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
		return string("ROLE_SUBCRIBER")
	case ROLE_PUBSUB:
		return string("ROLE_PUB_SUB")
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

func (params Parameters) Parse(reader quicvarint.Reader) error {

	length, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	for range length {
		ptype, err := quicvarint.Read(reader)

		if err != nil {
			return err
		}

		var param Parameter

		switch ptype {
		case ROLE_PARAM:
			param = &IntParameter{Ptype: ptype}
		case ROLE_PATH:
			param = &StringParameter{ptype: ptype}
		default:
			len, err := quicvarint.Read(reader)

			if err != nil {
				return err
			}

			discardData := make([]byte, len)
			reader.Read(discardData)

			return fmt.Errorf("[Unknown Param]")
		}

		err = param.Parse(reader)

		if err != nil {
			return err
		}

		params[ptype] = param
	}

	return nil
}

func (params Parameters) GetBytes() []byte {
	var data []byte
	paramlen := uint64(len(params))
	data = quicvarint.Append(data, paramlen)

	for _, param := range params {
		data = append(data, param.GetBytes()...)
	}

	return data
}

func (params Parameters) String() string {

	str := "[{"

	for _, param := range params {
		str += param.String() + " "
	}

	str += "}]"

	return str
}

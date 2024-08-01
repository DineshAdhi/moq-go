package moqt

import (
	"bufio"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	ROLE_PARAM = 0x00
	PATH_PARAM = 0x01
)

type Parameter interface {
	GetBytes() []byte
	Parse(r MOQTReader) error
	String() string
}

type Parameters map[uint64]Parameter

func GetParamKeyString(ptype uint64) string {
	switch ptype {
	case ROLE_PARAM:
		return "ROLE"
	case PATH_PARAM:
		return "PATH"
	default:
		return "UNKNOWN PARAM"
	}
}

func (params Parameters) Parse(r MOQTReader) error {

	reader := bufio.NewReader(r)
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
			intParam := IntParameter{ptype: ptype}
			err := intParam.Parse(r)

			if err != nil {
				return err
			}

			params[ptype] = intParam
		case PATH_PARAM:
			strParam := StringParameter{ptype: ptype}
			err := strParam.Parse(r)

			if err != nil {
				return err
			}

			params[ptype] = strParam
		default:
			len, err := quicvarint.Read(reader)
			if err != nil {
				return err
			}

			reader.Discard(int(len)) // Discarding unknown param
		}
	}

	return nil
}

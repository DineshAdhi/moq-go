package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type IntParameter struct {
	Ptype  uint64
	Pvalue uint64
}

func NewIntParameter(ptype uint64, pvalue uint64) *IntParameter {
	return &IntParameter{
		Ptype: ptype, Pvalue: pvalue,
	}
}

func (param IntParameter) Type() uint64 {
	return param.Ptype
}

func (param IntParameter) Value() interface{} {
	return param.Pvalue
}

func (param *IntParameter) Parse(reader quicvarint.Reader) error {

	_, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	pvalue, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	param.Pvalue = pvalue

	return nil
}

func (param IntParameter) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, param.Ptype)
	data = quicvarint.Append(data, uint64(quicvarint.Len(param.Pvalue)))
	data = quicvarint.Append(data, param.Pvalue)

	return data
}

func (param IntParameter) String() string {
	return fmt.Sprintf("%s : %s(0x%02X)", GetParamKeyString(&param), GetParamValueString(&param), param.Pvalue)
}

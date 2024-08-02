package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type IntParameter struct {
	ptype  uint64
	pvalue uint64
}

func (param IntParameter) Type() uint64 {
	return param.ptype
}

func (param IntParameter) Value() interface{} {
	return param.pvalue
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

	param.pvalue = pvalue

	return nil
}

func (param IntParameter) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, uint64(quicvarint.Len(param.pvalue)))
	data = quicvarint.Append(data, param.pvalue)

	return data
}

func (param IntParameter) String() string {
	return fmt.Sprintf("%s : %s(0x%02X)", GetParamKeyString(&param), GetParamValueString(&param), param.pvalue)
}

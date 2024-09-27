package wire

import (
	"fmt"
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

type StringParameter struct {
	ptype  uint64
	pvalue string
}

func ParseVarIntString(reader quicvarint.Reader) (string, error) {
	len, err := quicvarint.Read(reader)

	if err != nil {
		return "", err
	}

	data := make([]byte, len)
	n, err := io.ReadFull(reader, data)

	if err != nil {
		return "", err
	}

	str := string(data[:n])

	return str, nil
}

func GetBytesVarIntString(str string) []byte {
	var data []byte
	data = quicvarint.Append(data, uint64(len(str)))
	data = append(data, str...)

	return data
}

func (param StringParameter) Type() uint64 {
	return param.ptype
}

func (param StringParameter) Value() interface{} {
	return param.pvalue
}

func (param *StringParameter) Parse(reader quicvarint.Reader) error {
	str, err := ParseVarIntString(reader)

	if err != nil {
		return err
	}

	param.pvalue = str

	return nil
}

func (param StringParameter) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, param.ptype)
	data = quicvarint.Append(data, uint64(len(param.pvalue)))
	data = append(data, param.pvalue...)

	return data
}

func (param StringParameter) String() string {
	return fmt.Sprintf("Key : %v, Value : %v", GetParamKeyString(&param), param.pvalue)
}

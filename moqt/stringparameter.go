package moqt

import (
	"bufio"
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type StringParameter struct {
	ptype  uint64
	pvalue string
}

func (param StringParameter) Parse(r MOQTReader) error {

	reader := quicvarint.NewReader(r)
	len, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	data := make([]byte, len)
	reader = bufio.NewReader(r)
	reader.Read(data)

	if err != nil {
		return err
	}

	param.pvalue = string(data)

	return nil
}

func (param StringParameter) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, uint64(len(param.pvalue)))
	data = append(data, param.pvalue...)

	return data
}

func (param StringParameter) String() string {
	return fmt.Sprintf("Key : %v, Value : %v", param.ptype, param.pvalue)
}

package moqt

import (
	"bufio"

	"github.com/quic-go/quic-go/quicvarint"
)

type ServerSetup struct {
	SelectedVersion uint64
	Params          Parameters
}

func (setup *ServerSetup) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, setup.SelectedVersion)

	nparams := uint64(len(setup.Params))
	data = quicvarint.Append(data, nparams)

	for ptype, param := range setup.Params {
		data = quicvarint.Append(data, ptype)
		pvalue := param.GetBytes()
		data = append(data, pvalue...)
	}

	return data
}

func (setup *ServerSetup) Parse(r MOQTReader) error {

	reader := bufio.NewReader(r)
	var err error

	setup.SelectedVersion, err = quicvarint.Read(reader)

	if err != nil {
		return err
	}

	params := Parameters{}
	params.Parse(r)

	setup.Params = params

	return nil
}

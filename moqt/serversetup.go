package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type ServerSetup struct {
	SelectedVersion uint64
	Params          Parameters
}

func (setup ServerSetup) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, SERVER_SETUP)
	data = quicvarint.Append(data, setup.SelectedVersion)

	nparams := uint64(len(setup.Params))
	data = quicvarint.Append(data, nparams)

	for _, param := range setup.Params {
		data = append(data, param.GetBytes()...)
	}

	return data
}

func (setup *ServerSetup) Parse(reader quicvarint.Reader) error {

	var err error

	setup.SelectedVersion, err = quicvarint.Read(reader)

	if err != nil {
		return err
	}

	params := Parameters{}
	params.Parse(reader)

	setup.Params = params

	return nil
}

func (setup ServerSetup) String() string {

	str := fmt.Sprintf("[%s]", GetMoqMessageString(SERVER_SETUP))
	str += fmt.Sprintf("[Selected Version - %x][{", setup.SelectedVersion)

	for _, param := range setup.Params {
		str += fmt.Sprintf("%s ", param.String())
	}

	return str
}

package wire

import (
	"fmt"
	"strings"

	"github.com/quic-go/quic-go/quicvarint"
)

type ClientSetup struct {
	SupportedVersions []uint64
	Params            Parameters
}

func (setup ClientSetup) Type() uint64 {
	return CLIENT_SETUP
}

func (setup ClientSetup) GetRoleParam() (uint64, error) {
	param := setup.Params.GetParameter(ROLE_PARAM)

	if param != nil {
		return param.Value().(uint64), nil
	}

	return 0, fmt.Errorf("%s", "Param not found")
}

func (setup *ClientSetup) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, CLIENT_SETUP)
	nversions := uint64(len(setup.SupportedVersions))

	data = quicvarint.Append(data, nversions)

	for _, version := range setup.SupportedVersions {
		data = quicvarint.Append(data, version)
	}

	nparams := uint64(len(setup.Params))
	data = quicvarint.Append(data, nparams)

	for _, param := range setup.Params {
		pvalue := param.GetBytes()
		data = append(data, pvalue...)
	}

	return data
}

func (setup *ClientSetup) Parse(reader quicvarint.Reader) error {

	var err error

	nversions, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	for range nversions {
		version, err := quicvarint.Read(reader)

		if err != nil {
			return err
		}

		setup.SupportedVersions = append(setup.SupportedVersions, version)
	}

	params := Parameters{}
	params.Parse(reader)

	setup.Params = params

	return nil
}

func (setup ClientSetup) String() string {
	str := fmt.Sprintf("[%s]", GetMoqMessageString(CLIENT_SETUP))
	str += "[Supported Versions - {"

	for _, version := range setup.SupportedVersions {
		str += fmt.Sprintf("%X ", version)
	}

	str = strings.TrimSuffix(str, " ")

	str += "}][{"

	for _, param := range setup.Params {
		str += param.String() + ","
	}

	str = strings.TrimSuffix(str, ",")

	str += "}]"

	return str
}

func (setup ClientSetup) CheckDraftSupport() bool {

	for _, version := range setup.SupportedVersions {
		if version == DRAFT_04 {
			return true
		}
	}

	return false
}

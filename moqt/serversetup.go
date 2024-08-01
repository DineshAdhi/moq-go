package moqt

import "github.com/quic-go/quic-go/quicvarint"

type ServerSetup struct {
	SelectedVersion uint64
	Params          []SetupParameter
}

func (s *ServerSetup) GetBytes() []byte {
	data := []byte{}

	data = quicvarint.Append(data, uint64(SERVER_SETUP))      // MOQT Type
	data = quicvarint.Append(data, uint64(s.SelectedVersion)) // Supported Version

	length := len(s.Params)

	data = quicvarint.Append(data, uint64(length))

	for _, param := range s.Params {
		data = quicvarint.Append(data, param.ptype)
		var len uint64 = uint64(quicvarint.Len(param.pvalue))
		data = quicvarint.Append(data, len)

		data = quicvarint.Append(data, param.pvalue)
	}

	return data
}

func DefaultServerSetup() ServerSetup {
	s := ServerSetup{}

	s.SelectedVersion = DRAFT_03
	s.Params = []SetupParameter{
		{
			ptype:  ROLE_PARAM,
			pvalue: SUBSCRIBER,
		},
	}

	return s
}

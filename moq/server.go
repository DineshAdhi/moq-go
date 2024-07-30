package moq

import "github.com/quic-go/quic-go/quicvarint"

type ServerSetup struct {
	SelectedVersion uint64
	Params          []SetupParameter
}

func (s *ServerSetup) GetBytes() []byte {
	data := []byte{}

	quicvarint.Append(data, SERVER_SETUP)      // MOQT Type
	quicvarint.Append(data, s.SelectedVersion) // Supported Version

	for _, param := range s.Params {
		quicvarint.Append(data, param.ptype)
		var len uint64 = uint64(quicvarint.Len(param.pvalue))
		quicvarint.Append(data, len)

		quicvarint.Append(data, param.pvalue)
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

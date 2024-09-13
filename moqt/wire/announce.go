package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type Announce struct {
	TrackNameSpace string
	params         Parameters
}

func (a Announce) Type() uint64 {
	return ANNOUNCE
}

func (a *Announce) Parse(r quicvarint.Reader) (err error) {

	if a.TrackNameSpace, err = ParseVarIntString(r); err != nil {
		return
	}

	params := Parameters{}
	err = params.Parse(r)

	if err != nil {
		return
	}

	a.params = params

	return nil
}

func (a Announce) GetBytes() []byte {
	var data []byte

	namebytes := []byte(a.TrackNameSpace)
	data = quicvarint.Append(data, ANNOUNCE)
	data = quicvarint.Append(data, uint64(len(namebytes)))
	data = append(data, namebytes...)

	data = append(data, a.params.GetBytes()...)

	return data
}

func (a Announce) String() string {
	str := fmt.Sprintf("[ANNOUNCE][Track Namespace - %s]", a.TrackNameSpace)

	if len(a.params) > 0 {
		str += fmt.Sprintf("[Params - %s]", a.params.String())
	}

	return str
}

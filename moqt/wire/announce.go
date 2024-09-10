package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceMessage struct {
	TrackNameSpace string
	params         Parameters
}

func (a AnnounceMessage) Type() uint64 {
	return ANNOUNCE
}

func (a *AnnounceMessage) Parse(r quicvarint.Reader) error {

	namelen, err := quicvarint.Read(r)

	if err != nil {
		return err
	}

	namebytes := make([]byte, namelen)
	r.Read(namebytes)

	stringname := string(namebytes)

	a.TrackNameSpace = stringname

	params := Parameters{}
	err = params.Parse(r)

	if err != nil {
		return err
	}

	a.params = params

	return nil
}

func (a AnnounceMessage) GetBytes() []byte {
	var data []byte

	namebytes := []byte(a.TrackNameSpace)
	data = quicvarint.Append(data, ANNOUNCE)
	data = quicvarint.Append(data, uint64(len(namebytes)))
	data = append(data, namebytes...)

	data = append(data, a.params.GetBytes()...)

	return data
}

func (a AnnounceMessage) String() string {
	str := fmt.Sprintf("[ANNOUNCE][Track Namespace - %s]", a.TrackNameSpace)

	if len(a.params) > 0 {
		str += fmt.Sprintf("[Params - %s]", a.params.String())
	}

	return str
}

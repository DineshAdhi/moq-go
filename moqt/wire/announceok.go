package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type AnnounceOkMessage struct {
	TrackNameSpace string
}

func (msg AnnounceOkMessage) String() string {
	return fmt.Sprintf("[%s][Track Namespace - %s]", GetMoqMessageString(msg.Type()), msg.TrackNameSpace)
}

func (msg AnnounceOkMessage) Type() uint64 {
	return ANNOUNCE_OK
}

func (msg *AnnounceOkMessage) Parse(reader quicvarint.Reader) (err error) {
	msg.TrackNameSpace, err = ParseVarIntString(reader)
	return err
}

func (msg AnnounceOkMessage) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, ANNOUNCE_OK)
	data = quicvarint.Append(data, uint64(len(msg.TrackNameSpace)))
	data = append(data, msg.TrackNameSpace...)

	return data
}

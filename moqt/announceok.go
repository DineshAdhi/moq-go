package moqt

import "github.com/quic-go/quic-go/quicvarint"

type AnnounceOkMessage struct {
	tracknamespace string
}

func (msg *AnnounceOkMessage) Parse(reader quicvarint.Reader) (err error) {
	msg.tracknamespace, err = ParseVarIntString(reader)
	return err
}

func (msg AnnounceOkMessage) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, ANNOUNCE_OK)
	data = quicvarint.Append(data, uint64(len(msg.tracknamespace)))
	data = append(data, msg.tracknamespace...)

	return data
}

func (msg *AnnounceOkMessage) GetAnnounceOk(a *AnnounceMessage) {
	msg.tracknamespace = a.tracknamespace
}

package moqt

type PubStream struct {
	*MOQTSession
	StreamID   string
	SubID      uint64
	TrackAlias uint64
}

func NewPubStream(streamid string, subid uint64, alias uint64) *PubStream {
	return &PubStream{
		StreamID:   streamid,
		SubID:      subid,
		TrackAlias: alias,
	}
}

func (ps *PubStream) GetStreamID() string {
	return ps.StreamID
}

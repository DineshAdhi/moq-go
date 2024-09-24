package moqt

import "moq-go/moqt/wire"

type SubStream struct {
	*MOQTSession
	StreamId   string
	SubId      uint64
	TrackAlias uint64
	Namespace  string
	TrackName  string
	ObjectChan chan wire.TrackObject
}

func NewSubStream(session *MOQTSession, ns string, name string, streamid string, subid uint64, alias uint64) *SubStream {
	return &SubStream{
		MOQTSession: session,
		Namespace:   ns,
		TrackName:   name,
		StreamId:    streamid,
		SubId:       subid,
		TrackAlias:  alias,
		ObjectChan:  make(chan wire.TrackObject, 1024),
	}
}

func (ps SubStream) GetStreamID() string {
	return ps.StreamId
}

func (ps SubStream) GetSubID() uint64 {
	return ps.SubId
}

func (ps SubStream) GetAlias() uint64 {
	return ps.TrackAlias
}

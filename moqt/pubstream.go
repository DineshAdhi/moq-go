package moqt

import (
	"moq-go/moqt/wire"
)

type PubStream struct {
	*MOQTSession
	StreamId     string
	SubId        uint64
	TrackAlias   uint64
	Namespace    string
	TrackName    string
	GroupCounter uint64
}

func NewPubStream(session *MOQTSession, ns string, name string, streamid string, subid uint64, alias uint64) *PubStream {
	return &PubStream{
		MOQTSession:  session,
		Namespace:    ns,
		TrackName:    name,
		StreamId:     streamid,
		SubId:        subid,
		TrackAlias:   alias,
		GroupCounter: 0,
	}
}

func (ps PubStream) GetStreamID() string {
	return ps.StreamId
}

func (ps PubStream) GetSubID() uint64 {
	return ps.SubId
}

func (ps PubStream) GetAlias() uint64 {
	return ps.TrackAlias
}

func (ps PubStream) Push(objid uint64, data []byte) {

	unistream, err := ps.Conn.OpenUniStream()

	if err != nil {
		return
	}

	defer unistream.Close()

	header := wire.GroupHeader{
		SubscribeID: ps.SubId,
		TrackAlias:  ps.TrackAlias,
		GroupID:     ps.GroupCounter,
		SendOrder:   0,
	}

	ps.GroupCounter++

	unistream.Write(header.GetBytes(ps.SubId))

	object := wire.TrackObject{
		ObjectID: objid,
		Payload:  data,
	}

	unistream.Write(object.GetBytes())
}

package moqt

import (
	"fmt"
	"math/rand/v2"
	"moq-go/moqt/wire"

	"github.com/quic-go/quic-go/quicvarint"
)

type SubHandler struct {
	session           *MOQTSession
	SubscribedStreams *StreamsMap
	StreamsChan       chan *SubStream
}

func CreateNewSubHandler(session *MOQTSession) *SubHandler {
	handler := SubHandler{
		session:           session,
		SubscribedStreams: NewStreamsMap(session),
		StreamsChan:       make(chan *SubStream),
	}

	return &handler
}

func (sub *SubHandler) Subscribe(ns string, trackname string, alias uint64) {

	subid := uint64(rand.Uint32())
	streamid := fmt.Sprintf("%s_%s_%d", ns, trackname, alias)

	msg := wire.Subscribe{
		SubscribeID:    subid,
		TrackName:      trackname,
		TrackNameSpace: ns,
		TrackAlias:     0,
		FilterType:     wire.LatestGroup,
	}

	ss := NewSubStream(sub.session, ns, trackname, streamid, subid, alias)
	sub.SubscribedStreams.AddStream(subid, ss)

	sub.session.CS.WriteControlMessage(&msg)
}

func (pub *SubHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {
	pub.session.Slogger.Info().Msgf(msg.String())

	stream := pub.SubscribedStreams.SubIDGetStream(msg.SubscribeID)

	if stream != nil {
		pub.StreamsChan <- stream.(*SubStream)
	}
}

func (pub *SubHandler) HandleAnnounce(msg *wire.Announce) {

}

func (pub *SubHandler) HandleSubscribe(msg *wire.Subscribe) {
	pub.session.Close(wire.MOQERR_PROTOCOL_VIOLATION, "Protocol Violation. I am a Sub, dont send Subscribe to me.")
}

func (pub *SubHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	pub.session.Close(wire.MOQERR_PROTOCOL_VIOLATION, "Protocol Violation. I am a Sub, received announce_ok")
}

func (pub *SubHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

}

func (pub *SubHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {

}

func (pub *SubHandler) ProcessTracks() {

	s := pub.session

	for {
		unistream, err := s.Conn.AcceptUniStream(s.ctx)

		if err != nil {
			return
		}

		reader := quicvarint.NewReader(unistream)
		header, err := wire.ParseMOQTObjectHeader(reader)

		if err != nil {
			return
		}

		object := wire.TrackObject{}
		object.Parse(reader)

		subid := header.GetSubID()
		stream := pub.SubscribedStreams.SubIDGetStream(subid)

		ss := (stream).(*SubStream)

		ss.ObjectChan <- object
	}
}

package moqt

import (
	"math/rand"
	"moq-go/moqt/wire"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

type SubHandler struct {
	*MOQTSession
	SubscribedStreams StreamsMap[*SubStream]
	StreamsChan       chan SubStream
	AnnounceChan      chan string
}

func NewSubHandler(session *MOQTSession) *SubHandler {
	return &SubHandler{
		MOQTSession:       session,
		SubscribedStreams: NewStreamsMap[*SubStream](session),
		StreamsChan:       make(chan SubStream),
		AnnounceChan:      make(chan string),
	}
}

func (sub *SubHandler) Subscribe(ns string, name string, alias uint64) {

	subid := uint64(rand.Uint32())

	msg := wire.Subscribe{
		SubscribeID:    subid,
		TrackAlias:     alias,
		TrackName:      name,
		TrackNameSpace: ns,
		FilterType:     wire.LatestGroup,
	}

	sub.CS.WriteControlMessage(&msg)

	substream := NewSubStream(msg.GetStreamID(), subid)
	sub.SubscribedStreams.AddStream(subid, substream)
}

func (sub *SubHandler) HandleAnnounce(msg *wire.Announce) {
	log.Debug().Msg(msg.String())
	sub.AnnounceChan <- msg.TrackNameSpace
}

func (sub *SubHandler) HandleSubscribe(msg *wire.Subscribe) {
}

func (sub *SubHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {
	sub.Slogger.Info().Msg(msg.String())

	ss, ok := sub.SubscribedStreams.SubIDGetStream(msg.SubscribeID)

	if ok {
		sub.StreamsChan <- *ss
	} else {
		log.Error().Msgf("[Cannot find Substream with SubID - %X]", msg.SubscribeID)
	}
}

func (sub *SubHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {

}

func (sub *SubHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

}

func (sub *SubHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {

}

func (sub *SubHandler) DoHandle() {

	for {
		unistream, err := sub.Conn.AcceptUniStream(sub.ctx)

		if err != nil {
			sub.Slogger.Error().Msgf("[Error Accepting Unistream][%s]", err)
			return
		}

		reader := quicvarint.NewReader(unistream)
		subid, stream, err := wire.ParseMOQTStream(reader)

		if err != nil {
			sub.Slogger.Error().Msgf("[Error Parsing MOQT Stream][%s]", err)
			continue
		}

		if ss, ok := sub.SubscribedStreams.SubIDGetStream(subid); ok {
			go ss.ProcessObjects(stream, reader)
		} else {
			sub.Slogger.Error().Msgf("Received Header with unknown subid - %X", subid)
		}
	}
}

func (sub *SubHandler) HandleClose() {
	close(sub.StreamsChan)
}

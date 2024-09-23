package moqt

import (
	"math/rand/v2"
	"moq-go/moqt/wire"
	"strings"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

type RelayHandler struct {
	*MOQTSession
	IncomingStreams   StreamsMap
	SubscribedStreams StreamsMap
}

func CreateNewRelayHandler(session *MOQTSession) *RelayHandler {

	handler := &RelayHandler{
		MOQTSession: session,
	}

	handler.IncomingStreams = NewStreamsMap(session)
	handler.SubscribedStreams = NewStreamsMap(session)

	return handler
}

func (subscriber *RelayHandler) SendSubscribeOk(streamid string, okm *wire.SubscribeOk) {
	if subid, err := subscriber.SubscribedStreams.GetSubID(streamid); err == nil {
		subokmsg := *okm
		subokmsg.SubscribeID = subid

		subscriber.CS.WriteControlMessage(&subokmsg)
	}
}

// For publishers, a send subscribe will always yield a ObjectStream
func (publisher *RelayHandler) SendSubscribe(msg wire.Subscribe) *RelayObjectStream {

	streamid := msg.GetStreamID()
	subid := uint64(rand.Uint32())

	rs := NewRelayObjectStream(subid, streamid, &publisher.IncomingStreams, publisher.MOQTSession)
	publisher.IncomingStreams.AddStream(subid, streamid, rs)

	msg.SubscribeID = subid
	publisher.CS.WriteControlMessage(&msg)

	return rs
}

// Fetches the Object Stream with the StreamID (OR) Forwards the Subscribe and returns the RelayObjectStream Placeholder
func (publisher *RelayHandler) GetRelayObjectStream(msg *wire.Subscribe) *RelayObjectStream {

	streamid := msg.GetStreamID()
	stream, found := publisher.IncomingStreams.StreamIDGetStream(streamid)

	var rs *RelayObjectStream

	// We need to fetch the fresh copies of ".catalog", "audio.mp4", "video.mp4".I knowm its a nasty implementation. Requires more work.
	if !found || strings.Contains(msg.TrackName, ".catalog") || strings.Contains(msg.TrackName, ".mp4") {
		rs = publisher.SendSubscribe(*msg)
	} else {
		rs = stream.(*RelayObjectStream)
	}

	return rs
}

func (subscriber *RelayHandler) DispatchObject(object *MOQTObject) {

	if subid, err := subscriber.SubscribedStreams.GetSubID(object.GetStreamID()); err == nil {

		unistream, err := subscriber.Conn.OpenUniStream()

		if err != nil {
			// s.Slogger.Error().Msgf("[Error Opening Unistream][%s]", err)
			return
		}

		defer unistream.Close()

		groupHeader := object.Header
		unistream.Write(groupHeader.GetBytes(subid))

		reader := object.NewReader()
		reader.Pipe(unistream)
	} else {
		subscriber.Slogger.Error().Msgf("[Unable to find DownStream SubID for StreamID][Stream ID - %s]", object.GetStreamID())
	}
}

func (publisher *RelayHandler) ProcessTracks() {

	for {
		unistream, err := publisher.Conn.AcceptUniStream(publisher.ctx)

		if err != nil {
			log.Error().Msgf("[Error Acceping Unistream][%s]", err)
			break
		}

		reader := quicvarint.NewReader(unistream)
		header, err := wire.ParseMOQTObjectHeader(reader)

		if err != nil {
			publisher.Slogger.Error().Msgf("[Error Parsing Object Header][%s]", err)
			continue
		}

		subid := header.GetSubID()

		if rs, found := publisher.IncomingStreams.SubIDGetStream(subid).(*RelayObjectStream); found {

			object := NewMOQTObject(header, rs.streamid, reader)
			rs.AddObject(object)

		} else {
			publisher.Slogger.Error().Msgf("[Object Stream Not Found][Alias - %d]", header.GetTrackAlias())
		}
	}
}

// Comes from Publisher
func (publisher *RelayHandler) HandleAnnounce(msg *wire.Announce) {

	publisher.Slogger.Info().Msgf(msg.String())

	okmsg := wire.AnnounceOk{}
	okmsg.TrackNameSpace = msg.TrackNameSpace

	sm.addPublisher(msg.TrackNameSpace, publisher)

	publisher.CS.WriteControlMessage(&okmsg)
}

// Comes from Publisher
func (publisher *RelayHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {
	publisher.Slogger.Info().Msg(msg.String())

	subid := msg.SubscribeID

	if rs, ok := publisher.IncomingStreams.SubIDGetStream(subid).(*RelayObjectStream); ok {

		for _, sub := range rs.subscribers {
			sub.SendSubscribeOk(rs.streamid, msg)
		}
	}
}

func (publisher *RelayHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {
	publisher.Slogger.Info().Msg(msg.String())
}

// Comes from Subscriber
func (subscriber *RelayHandler) HandleSubscribe(msg *wire.Subscribe) {

	subscriber.Slogger.Info().Msg(msg.String())

	pub := sm.getPublisher(msg.TrackNameSpace)

	if pub == nil {
		log.Error().Msgf("[No Publisher found with Namespace - %s]", msg.TrackNameSpace)
		return
	}

	os := pub.GetRelayObjectStream(msg)
	os.AddSubscriber(msg.SubscribeID, subscriber.MOQTSession)

	subscriber.SubscribedStreams.AddStream(msg.SubscribeID, msg.GetStreamID(), os)
}

func (subscriber *RelayHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	subscriber.Slogger.Info().Msg(msg.String())
}

func (subscriber *RelayHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

	subscriber.Slogger.Info().Msg(msg.String())

	subid := msg.SubscriptionID

	if rs, ok := subscriber.SubscribedStreams.SubIDGetStream(subid).(*RelayObjectStream); ok {
		rs.RemoveSubscriber(subscriber.Id)
	}
}

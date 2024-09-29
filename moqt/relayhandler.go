package moqt

import (
	"io"
	"math/rand/v2"
	"moq-go/moqt/wire"
	"strings"
	"sync"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

type RelayHandler struct {
	*MOQTSession
	IncomingStreams   StreamsMap[*RelayStream]
	SubscribedStreams StreamsMap[*RelayStream]
}

func CreateNewRelayHandler(session *MOQTSession) *RelayHandler {

	handler := &RelayHandler{
		MOQTSession: session,
	}

	handler.IncomingStreams = NewStreamsMap[*RelayStream](session)
	handler.SubscribedStreams = NewStreamsMap[*RelayStream](session)

	return handler
}

func (rh *RelayHandler) HandleClose() {
	for _, rs := range rh.SubscribedStreams.streams {
		rs.RemoveSubscriber(rh.Id)
	}
}

func (subscriber *RelayHandler) SendSubscribeOk(streamid string, okm wire.SubscribeOk) {
	if subid, err := subscriber.SubscribedStreams.GetSubID(streamid); err == nil {
		okm.SubscribeID = subid
		subscriber.CS.WriteControlMessage(&okm)
	}
}

// For publishers, a send subscribe will always yield a ObjectStream
func (publisher *RelayHandler) SendSubscribe(msg wire.Subscribe) *RelayStream {

	streamid := msg.GetStreamID()
	subid := uint64(rand.Uint32())

	rs := NewRelayStream(subid, streamid, &publisher.IncomingStreams)
	publisher.IncomingStreams.AddStream(subid, rs)

	publisher.Slogger.Info().Msgf("[New Incoming Stream][SubID - %x][Stream ID - %s]", subid, streamid)

	msg.SubscribeID = subid
	publisher.CS.WriteControlMessage(&msg)

	return rs
}

// Fetches the Object Stream with the StreamID (OR) Forwards the Subscribe and returns the ObjectStream Placeholder
func (publisher *RelayHandler) GetObjectStream(msg *wire.Subscribe) *RelayStream {

	streamid := msg.GetStreamID()
	stream := publisher.IncomingStreams.StreamIDGetStream(streamid)

	var rs *RelayStream

	// We need to fetch the fresh copies of ".catalog", "audio.mp4", "video.mp4".I knowm its a nasty implementation. Requires more work.
	if stream == nil || strings.Contains(msg.TrackName, ".catalog") || strings.Contains(msg.TrackName, ".mp4") {
		rs = publisher.SendSubscribe(*msg)
	} else {
		rs = stream.(*RelayStream)
	}

	return rs
}

func (sub *RelayHandler) ProcessMOQTStream(stream wire.MOQTStream, wg *sync.WaitGroup) {

	streamid := stream.GetStreamID()
	rs := sub.SubscribedStreams.StreamIDGetStream(streamid)

	if rs == nil {
		log.Error().Msgf("[Cannot find SubID for Stream - %s]", streamid)
		wg.Done()
		return
	}

	subid := rs.GetSubID()

	unistream, err := sub.Conn.OpenUniStream()

	if err != nil {
		sub.Slogger.Error().Msgf("[Cannot Open Unistream - %s]", err)
		wg.Done()
		return
	}

	unistream.Write(stream.GetHeaderSubIDBytes(subid))
	wg.Done()

	itr := 0

	for {
		itr, err = stream.Pipe(itr, unistream)

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Error().Msgf("[Error Piping Stream to Unistream][%s]", err)
			break
		}
	}

	unistream.Close()
}

func (publisher *RelayHandler) ProcessObjectStreams() {

	for {
		unistream, err := publisher.Conn.AcceptUniStream(publisher.ctx)

		if err != nil {
			publisher.Slogger.Error().Msgf("[Error Accepting Unistream][%s]", err)
			return
		}

		reader := quicvarint.NewReader(unistream)

		htype, err := quicvarint.Read(reader)

		if err != nil {
			publisher.Slogger.Error().Msgf("[Unable to read header type][%s]", err)
			return
		}

		switch htype {
		case wire.STREAM_HEADER_GROUP, wire.STREAM_HEADER_TRACK:
			break
		default:
			continue
		}

		subid, err := quicvarint.Read(reader)

		if err != nil {
			return
		}

		if rs := publisher.IncomingStreams.SubIDGetStream(subid); rs != nil {
			go rs.ProcessObjects(htype, subid, reader)
		} else {
			log.Error().Msgf("[Stream not found for SubID - %X]", subid)
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

	if rs := publisher.IncomingStreams.SubIDGetStream(subid); rs != nil {
		rs.ForwardSubscribeOk(*msg)
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

	rs := pub.GetObjectStream(msg)

	if rs == nil {
		log.Error().Msgf("[Object Stream not found][%s]", msg.GetStreamID())
	}

	rs.AddSubscriber(subscriber)

	subscriber.Slogger.Info().Msgf("[Subscribed to Stream - %s]", rs.GetStreamID())
	subscriber.SubscribedStreams.AddStream(msg.SubscribeID, rs)
}

func (subscriber *RelayHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	subscriber.Slogger.Info().Msg(msg.String())
}

func (subscriber *RelayHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

	subscriber.Slogger.Info().Msg(msg.String())

	subid := msg.SubscriptionID

	if rs := subscriber.SubscribedStreams.SubIDGetStream(subid); rs != nil {
		rs.RemoveSubscriber(subscriber.Id)
	}
}

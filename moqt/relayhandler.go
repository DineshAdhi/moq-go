package moqt

import (
	"fmt"
	"math/rand/v2"
	"moq-go/moqt/wire"
	"strings"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

type RelayHandler struct {
	*MOQTSession
	IncomingStreams   *StreamsMap
	SubscribedStreams *StreamsMap
	ObjectStreams     map[string]quic.SendStream
	ObjectStreamsLock sync.RWMutex
	ObjectChan        chan *wire.TrackObject
}

func CreateNewRelayHandler(session *MOQTSession) *RelayHandler {

	handler := &RelayHandler{
		MOQTSession: session,
	}

	handler.IncomingStreams = NewStreamsMap(session)
	handler.SubscribedStreams = NewStreamsMap(session)
	handler.ObjectStreams = map[string]quic.SendStream{}
	handler.ObjectStreamsLock = sync.RWMutex{}

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

	rs := NewRelayObjectStream(subid, streamid, msg.TrackAlias, publisher.IncomingStreams, publisher.MOQTSession)
	publisher.IncomingStreams.AddStream(subid, rs)

	msg.SubscribeID = subid
	publisher.CS.WriteControlMessage(&msg)

	return rs
}

// Fetches the Object Stream with the StreamID (OR) Forwards the Subscribe and returns the RelayObjectStream Placeholder
func (publisher *RelayHandler) GetRelayObjectStream(msg *wire.Subscribe) *RelayObjectStream {

	streamid := msg.GetStreamID()
	stream := publisher.IncomingStreams.StreamIDGetStream(streamid)

	var rs *RelayObjectStream

	// We need to fetch the fresh copies of ".catalog", "audio.mp4", "video.mp4".I knowm its a nasty implementation. Requires more work.
	if stream == nil || strings.Contains(msg.TrackName, ".catalog") || strings.Contains(msg.TrackName, ".mp4") {
		rs = publisher.SendSubscribe(*msg)
	} else {
		rs = stream.(*RelayObjectStream)
	}

	return rs
}

func (subscriber *RelayHandler) CreateObjectSream(header wire.MOQTObjectHeader, streamid string) (quic.SendStream, error) {
	subscriber.ObjectStreamsLock.Lock()
	defer subscriber.ObjectStreamsLock.Unlock()

	groupKey := header.GetGroupKey()

	if stream, ok := subscriber.ObjectStreams[groupKey]; ok {
		return stream, nil
	} else {
		unistream, err := subscriber.Conn.OpenUniStream()

		if err != nil {
			return nil, err
		}

		if subid, err := subscriber.SubscribedStreams.GetSubID(streamid); err == nil {

			unistream.Write(header.GetBytes(subid))
			subscriber.ObjectStreams[groupKey] = unistream

			log.Debug().Msgf("New Outgoing Object STream - %s", groupKey)
			return unistream, nil

		} else {
			return nil, fmt.Errorf("[Unable to find DownStream SubID for StreamID][Stream ID - %s]", streamid)
		}
	}
}

func (subscriber *RelayHandler) DeleteObjectStream(header wire.MOQTObjectHeader) {
	subscriber.ObjectStreamsLock.Lock()
	defer subscriber.ObjectStreamsLock.Unlock()

	groupKey := header.GetGroupKey()

	if stream, ok := subscriber.ObjectStreams[groupKey]; ok {
		stream.Close()
		delete(subscriber.ObjectStreams, groupKey)
		log.Debug().Msgf("Deleting Object STream - %s", groupKey)
	}
}

func (subscriber *RelayHandler) DispatchObject(obj *wire.TrackObject) {

	subscriber.ObjectStreamsLock.RLock()
	defer subscriber.ObjectStreamsLock.RUnlock()

	groupKey := obj.Header.GetGroupKey()

	if unistream, ok := subscriber.ObjectStreams[groupKey]; ok {
		data := obj.GetBytes()
		unistream.Write(data)
	} else {
		log.Debug().Msgf("Unable to Dispatch %s", groupKey)
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

		publisher.Slogger.Debug().Msgf("New Incoming Stream - %s", header.GetGroupKey())

		subid := header.GetSubID()

		if rs, found := publisher.IncomingStreams.SubIDGetStream(subid).(*RelayObjectStream); found {

			go rs.ProcessObjects(unistream, header)

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

	subscriber.SubscribedStreams.AddStream(msg.SubscribeID, os)
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

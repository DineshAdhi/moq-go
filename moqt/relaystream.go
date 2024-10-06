package moqt

import (
	"io"
	"sync"

	"github.com/DineshAdhi/moq-go/moqt/wire"

	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

type RelayStream struct {
	SubID           uint64
	StreamID        string
	Map             *StreamsMap[*RelayStream]
	Subscribers     map[string]*RelayHandler
	SubscribersLock sync.RWMutex
	ObjectCache     []wire.Object
}

func (rs *RelayStream) GetSubID() uint64 {
	return rs.SubID
}

func (rs *RelayStream) GetStreamID() string {
	return rs.StreamID
}

func NewRelayStream(subid uint64, id string, smap *StreamsMap[*RelayStream]) *RelayStream {

	rs := &RelayStream{}
	rs.SubID = subid
	rs.StreamID = id
	rs.Map = smap
	rs.Subscribers = map[string]*RelayHandler{}
	rs.SubscribersLock = sync.RWMutex{}
	rs.ObjectCache = make([]wire.Object, 0)

	return rs
}

func (rs *RelayStream) AddSubscriber(handler *RelayHandler) {
	rs.SubscribersLock.Lock()
	defer rs.SubscribersLock.Unlock()

	rs.Subscribers[handler.Id] = handler
}

func (os *RelayStream) RemoveSubscriber(id string) {
	os.SubscribersLock.Lock()
	defer os.SubscribersLock.Unlock()

	delete(os.Subscribers, id)
}

func (rs *RelayStream) ForwardSubscribeOk(msg wire.SubscribeOk) {

	for _, sub := range rs.Subscribers {
		if handler := sub.RelayHandler(); handler != nil {
			handler.SendSubscribeOk(rs.GetStreamID(), msg)
		}
	}
}

func (rs *RelayStream) ForwardStream(stream wire.MOQTStream) {
	rs.SubscribersLock.RLock()
	defer rs.SubscribersLock.RUnlock()

	for _, sub := range rs.Subscribers {
		go sub.ProcessMOQTStream(stream)
	}
}

func (rs *RelayStream) ProcessObjects(stream wire.MOQTStream, reader quicvarint.Reader) {

	rs.ForwardStream(stream)

	for {
		_, object, err := stream.ReadObject()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Debug().Msgf("[Error Reading Object][%s]", err)
			return
		}

		rs.ObjectCache = append(rs.ObjectCache, *object)
	}
}

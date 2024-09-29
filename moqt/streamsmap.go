package moqt

import (
	"fmt"
	"moq-go/moqt/wire"
	"sync"

	"github.com/quic-go/quic-go/quicvarint"
)

type ObjectStream interface {
	GetStreamID() string
	GetSubID() uint64
	AddSubscriber(handler *RelayHandler)
	ForwardSubscribeOk(wire.SubscribeOk)
	RemoveSubscriber(string)
	ProcessObjects(uint64, uint64, quicvarint.Reader)
}

// A simple Map util to keep ObjectStream of the ObjectStream respecitve to its streamid and subid
type StreamsMap[T ObjectStream] struct {
	*MOQTSession
	streams map[uint64]T // SubID - ObjectStream
	lock    sync.RWMutex
}

func NewStreamsMap[T ObjectStream](s *MOQTSession) StreamsMap[T] {
	return StreamsMap[T]{
		MOQTSession: s,
		streams:     map[uint64]T{},
		lock:        sync.RWMutex{},
	}
}

func (s *StreamsMap[T]) GetSubID(streamid string) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for subid, stream := range s.streams {
		if stream.GetStreamID() == streamid {
			return subid, nil
		}
	}

	return 0, fmt.Errorf("[SubID not found]")
}

func (s *StreamsMap[T]) StreamIDGetStream(streamid string) ObjectStream {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, stream := range s.streams {
		if stream.GetStreamID() == streamid {
			return stream
		}
	}

	return nil
}

func (s *StreamsMap[T]) SubIDGetStream(subid uint64) ObjectStream {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if stream, ok := s.streams[subid]; ok {
		return stream
	}

	return nil
}

func (s *StreamsMap[T]) AddStream(subid uint64, os T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.streams[subid] = os
}

func (s *StreamsMap[T]) DeleteStream(streamid string) {

	if stream := s.StreamIDGetStream(streamid); stream != nil {
		subid := stream.GetSubID()

		s.lock.Lock()
		defer s.lock.Unlock()

		s.Slogger.Debug().Msgf("[Deleting Stream - %s]", streamid)

		delete(s.streams, subid)
	}
}

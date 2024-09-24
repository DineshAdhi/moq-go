package moqt

import (
	"fmt"
	"sync"
)

// A simple Map util to keep track of the ObjectStream respecitve to its streamid and subid

type ObjectStream interface {
	GetStreamID() string
	GetSubID() uint64
	GetAlias() uint64
}

type StreamsMap struct {
	*MOQTSession
	streams map[uint64]ObjectStream // SubID - ObjectStream
	lock    sync.RWMutex
}

func NewStreamsMap(s *MOQTSession) *StreamsMap {
	return &StreamsMap{
		MOQTSession: s,
		streams:     map[uint64]ObjectStream{},
		lock:        sync.RWMutex{},
	}
}

func (s *StreamsMap) GetSubID(streamid string) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for subid, stream := range s.streams {
		if stream.GetStreamID() == streamid {
			return subid, nil
		}
	}

	return 0, fmt.Errorf("[SubID not found]")
}

func (s *StreamsMap) GetAlias(streamid string) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, stream := range s.streams {
		if stream.GetStreamID() == streamid {
			return stream.GetAlias(), nil
		}
	}

	return 0, fmt.Errorf("[Alias not found]")
}

func (s *StreamsMap) StreamIDGetStream(streamid string) ObjectStream {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, stream := range s.streams {
		if stream.GetStreamID() == streamid {
			return stream
		}
	}

	return nil
}

func (s *StreamsMap) SubIDGetStream(subid uint64) ObjectStream {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if stream, ok := s.streams[subid]; ok {
		return stream
	}

	return nil
}

func (s *StreamsMap) AddStream(subid uint64, os ObjectStream) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.streams[subid] = os
}

func (s *StreamsMap) DeleteStream(streamid string) {

	if stream := s.StreamIDGetStream(streamid); stream != nil {
		subid := stream.GetSubID()

		s.lock.Lock()
		defer s.lock.Unlock()

		s.Slogger.Debug().Msgf("[Deleting Stream - %s]", streamid)

		delete(s.streams, subid)
	}
}

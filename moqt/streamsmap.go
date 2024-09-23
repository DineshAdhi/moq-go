package moqt

import (
	"fmt"
	"sync"
)

// A simple Map util to keep track of the ObjectStream respecitve to its streamid and subid

type ObjectStream interface{}

type StreamsMap struct {
	*MOQTSession
	streams     map[uint64]ObjectStream // SubID - ObjectStream
	streamidmap map[string]uint64       // StreamID - SubID
	lock        sync.RWMutex
}

func NewStreamsMap(s *MOQTSession) StreamsMap {
	return StreamsMap{
		MOQTSession: s,
		streams:     map[uint64]ObjectStream{},
		streamidmap: map[string]uint64{},
		lock:        sync.RWMutex{},
	}
}

func (s *StreamsMap) GetSubID(streamid string) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if subid, ok := s.streamidmap[streamid]; ok {
		return subid, nil
	}

	return 0, fmt.Errorf("[SubID not found]")
}

func (s *StreamsMap) StreamIDGetStream(streamid string) (ObjectStream, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if subid, ok := s.streamidmap[streamid]; ok {
		if stream, ok := s.streams[subid]; ok {
			return stream, true
		}
	}

	return nil, false
}

func (s *StreamsMap) SubIDGetStream(subid uint64) ObjectStream {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if stream, ok := s.streams[subid]; ok {
		return stream
	}

	return nil
}

func (s *StreamsMap) AddStream(subid uint64, streamid string, os ObjectStream) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.streamidmap[streamid] = subid
	s.streams[subid] = os
}

func (s *StreamsMap) DeleteStream(streamid string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if subid, found := s.streamidmap[streamid]; found {
		delete(s.streams, subid)
		delete(s.streamidmap, streamid)

		s.Slogger.Info().Msgf("[Deleting Stream][%s]", streamid)
	}
}

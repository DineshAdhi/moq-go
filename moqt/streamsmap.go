package moqt

import (
	"sync"
)

type StreamsMap struct {
	*MOQTSession
	streams  map[uint64]*ObjectStream // SubID - ObjectStream
	subidmap map[uint64]uint64        // StreamID - SubID
	lock     sync.RWMutex
}

func NewStreamsMap(s *MOQTSession) StreamsMap {
	return StreamsMap{
		MOQTSession: s,
		streams:     map[uint64]*ObjectStream{},
		subidmap:    map[uint64]uint64{},
		lock:        sync.RWMutex{},
	}
}

func (s *StreamsMap) StreamIDGetStream(id uint64) (*ObjectStream, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if subid, ok := s.subidmap[id]; ok {
		if stream, ok := s.streams[subid]; ok {
			return stream, true
		}
	}

	return nil, false
}

func (s *StreamsMap) SubIDGetStream(subid uint64) (*ObjectStream, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if stream, ok := s.streams[subid]; ok {
		return stream, true
	}

	return nil, false
}

func (s *StreamsMap) AddStream(subid uint64, os *ObjectStream) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.subidmap[os.streamid] = subid
	s.streams[subid] = os
}

func (s *StreamsMap) DeleteStream(os *ObjectStream) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Slogger.Info().Msgf("[Deleting Stream][%d]", os.streamid)

	delete(s.subidmap, os.streamid)
	delete(s.streams, os.subid)
}

func (s *StreamsMap) CreateNewStream(subid uint64, streamid uint64) *ObjectStream {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.subidmap[streamid] = subid

	obs := NewObjectStream(subid, streamid, s)
	s.streams[subid] = obs

	s.Slogger.Info().Msgf("[New Object Stream][%d]", streamid)

	return obs
}

func (s *StreamsMap) DeleteAll() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, os := range s.streams {
		os.RemoveSubscriber(s.id)
	}
}
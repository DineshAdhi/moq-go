package moqt

import (
	"sync"
)

type StreamsMap struct {
	*MOQTSession
	streams     map[uint64]*ObjectStream // SubID - ObjectStream
	streamidmap map[string]uint64        // StreamID - SubID
	lock        sync.RWMutex
}

func NewStreamsMap(s *MOQTSession) StreamsMap {
	return StreamsMap{
		MOQTSession: s,
		streams:     map[uint64]*ObjectStream{},
		streamidmap: map[string]uint64{},
		lock:        sync.RWMutex{},
	}
}

func (s *StreamsMap) StreamIDGetStream(streamid string) (*ObjectStream, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if subid, ok := s.streamidmap[streamid]; ok {
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

	s.streamidmap[os.streamid] = subid
	s.streams[subid] = os
}

func (s *StreamsMap) DeleteStream(os *ObjectStream) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Slogger.Info().Msgf("[Deleting Stream][%s]", os.streamid)

	delete(s.streamidmap, os.streamid)
	delete(s.streams, os.subid)

	if s.isUpstream() {

		// Send Unsubscribe to Upstream
		// s.SendUnsubscribe(os.subid) moq-js doesn't support it yet

		// Proceed deleting the streams to downstream subscribers
		for _, sub := range os.subscribers {
			sub.SubscribedStreams.DeleteStream(os)
		}
	}
}

func (s *StreamsMap) CreateNewStream(subid uint64, streamid string) *ObjectStream {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.streamidmap[streamid] = subid

	obs := NewObjectStream(subid, streamid, s)
	s.streams[subid] = obs

	s.Slogger.Info().Msgf("[New Object Stream][%s]", streamid)

	return obs
}

func (s *StreamsMap) DeleteAll() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, os := range s.streams {
		os.RemoveSubscriber(s.id)
	}
}

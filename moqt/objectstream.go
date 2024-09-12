package moqt

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type ObjectStream struct {
	streamsmap     *StreamsMap
	subid          uint64
	streamid       string
	subscribers    map[string]*MOQTSession
	subscriberlock sync.RWMutex
	objectlock     sync.RWMutex
	objects        map[string]*MOQTObject
}

func NewObjectStream(subid uint64, streamid string, sm *StreamsMap) *ObjectStream {

	os := &ObjectStream{
		streamsmap:     sm,
		subid:          subid,
		streamid:       streamid,
		subscriberlock: sync.RWMutex{},
		objectlock:     sync.RWMutex{},
		subscribers:    map[string]*MOQTSession{},
		objects:        map[string]*MOQTObject{},
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		quit := make(chan struct{})
		for {
			select {
			case <-ticker.C:
				os.CleanUp(quit)
			case <-quit:
				ticker.Stop()
				break
			}
		}
	}()

	return os
}

func (os *ObjectStream) CleanUp(quit chan struct{}) {
	os.objectlock.Lock()
	defer os.objectlock.Unlock()

	expiredlist := []string{}

	for id, obj := range os.objects {
		if obj.isExpired() {
			expiredlist = append(expiredlist, id)
		}
	}

	for _, expired := range expiredlist {
		delete(os.objects, expired)
	}

	if len(os.objects) == 0 {
		os.streamsmap.DeleteStream(os)
		close(quit)
	}
}

func (os *ObjectStream) AddSubscriber(subid uint64, session *MOQTSession) {
	os.subscriberlock.Lock()
	defer os.subscriberlock.Unlock()

	os.subscribers[session.id] = session

	session.Slogger.Info().Msgf("[Subscribed to Stream - %s][Len - %d]", os.streamid, len(os.subscribers))
}

func (os *ObjectStream) RemoveSubscriber(sessionid string) {
	os.subscriberlock.Lock()
	defer os.subscriberlock.Unlock()

	log.Debug().Msgf("[Session Unsubscribed from - %s][Session - %s]", os.streamid, sessionid)

	delete(os.subscribers, sessionid)
}

func (os *ObjectStream) AddObject(object *MOQTObject) {
	os.objectlock.Lock()
	defer os.objectlock.Unlock()

	object.SetStreamID(os.streamid) // Very Important. Object only contains Alias. Set the StreamID, so its easy for downstream to get the subid
	os.objects[object.header.GetObjectKey()] = object

	go os.NotifySubscribers(object)
}

func (os *ObjectStream) NotifySubscribers(object *MOQTObject) {
	os.subscriberlock.RLock()
	defer os.subscriberlock.RUnlock()

	for _, subscriber := range os.subscribers {
		go subscriber.DispatchObject(object)
	}
}

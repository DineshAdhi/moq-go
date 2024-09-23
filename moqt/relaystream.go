package moqt

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	CLEAN_UP_INTERVAL = 20
)

type RelayObjectStream struct {
	*MOQTSession
	streamsmap     *StreamsMap
	subid          uint64
	streamid       string
	subscribers    map[string]*RelayHandler
	subscriberlock sync.RWMutex
	objectlock     sync.RWMutex
	objects        map[string]*MOQTObject
	stopCleanup    bool
}

func NewRelayObjectStream(subid uint64, streamid string, sm *StreamsMap, session *MOQTSession) *RelayObjectStream {

	os := RelayObjectStream{
		MOQTSession:    session,
		streamsmap:     sm,
		subid:          subid,
		streamid:       streamid,
		subscriberlock: sync.RWMutex{},
		objectlock:     sync.RWMutex{},
		subscribers:    map[string]*RelayHandler{},
		objects:        map[string]*MOQTObject{},
		stopCleanup:    false,
	}

	go func() {
		ticker := time.NewTicker(time.Second * CLEAN_UP_INTERVAL)

		closechannel := make(chan bool)

		for {
			select {
			case <-ticker.C:
				os.CleanUp(closechannel)
			case <-closechannel:
				ticker.Stop()
				return
			}
		}
	}()

	return &os
}

func (os *RelayObjectStream) CleanUp(closechannel chan bool) bool {
	os.objectlock.Lock()
	defer os.objectlock.Unlock()

	expiredlist := []string{}

	for id, obj := range os.objects {
		if obj.IsExpired() {
			expiredlist = append(expiredlist, id)
		}
	}

	for _, expired := range expiredlist {
		delete(os.objects, expired)
	}

	if len(os.objects) == 0 {
		go os.DeleteStream()
		close(closechannel)
	}

	return false
}

func (os *RelayObjectStream) DeleteStream() {

	os.streamsmap.DeleteStream(os.streamid)

	for _, sub := range os.subscribers {
		sub.SubscribedStreams.DeleteStream(os.streamid)
	}
}

func (os *RelayObjectStream) AddSubscriber(subid uint64, session *MOQTSession) {
	os.subscriberlock.Lock()
	defer os.subscriberlock.Unlock()

	os.subscribers[session.Id] = session.Handler.(*RelayHandler)

	session.Slogger.Info().Msgf("[Subscribed to Stream - %s][Len - %d]", os.streamid, len(os.subscribers))
}

func (os *RelayObjectStream) RemoveSubscriber(sessionid string) {
	os.subscriberlock.Lock()
	defer os.subscriberlock.Unlock()

	log.Debug().Msgf("[Session Unsubscribed from - %s][%s]", os.streamid, sessionid)

	delete(os.subscribers, sessionid)
}

func (os *RelayObjectStream) AddObject(object *MOQTObject) {
	os.objectlock.Lock()
	os.objects[object.Header.GetObjectKey()] = object
	os.objectlock.Unlock()

	go object.ParseFromStream(object.Reader)

	os.NotifySubscribers(object)
}

func (os *RelayObjectStream) NotifySubscribers(object *MOQTObject) {
	os.subscriberlock.RLock()
	defer os.subscriberlock.RUnlock()

	for _, subscriber := range os.subscribers {
		go subscriber.DispatchObject(object)
	}
}

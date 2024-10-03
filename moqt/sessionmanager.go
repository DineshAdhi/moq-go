package moqt

import (
	"sync"
	"time"

	"github.com/DineshAdhi/moq-go/moqt/wire"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/rand"
)

type SessionManager struct {
	sessionslock sync.RWMutex
	nslock       sync.RWMutex
	cachelock    sync.RWMutex
	sessions     map[string]*MOQTSession
	namespaces   map[string]*RelayHandler
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{}
	sm.sessionslock = sync.RWMutex{}
	sm.nslock = sync.RWMutex{}
	sm.namespaces = map[string]*RelayHandler{}
	sm.sessions = map[string]*MOQTSession{}

	rand.Seed(uint64(time.Now().UnixNano()))

	return sm
}

func (sm *SessionManager) addSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	sm.sessions[session.Id] = session

	log.Info().Msgf("[%s][New MOQT Session]", session.Id)
}

func (sm *SessionManager) removeSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	id := session.Id
	delete(sm.sessions, id)
}

func (sm *SessionManager) addPublisher(ns string, pub *RelayHandler) {
	sm.nslock.Lock()
	sm.namespaces[ns] = pub
	sm.nslock.Unlock()

	sm.sessionslock.RLock()

	for _, peer := range sm.sessions {

		if (peer.RemoteRole == wire.ROLE_RELAY || peer.RemoteRole == wire.ROLE_SUBSCRIBER) && peer.Id != pub.Id {
			announce := wire.Announce{
				TrackNameSpace: ns,
			}

			go peer.CS.WriteControlMessage(&announce)
		}
	}

	sm.sessionslock.RUnlock()
}

func (sm *SessionManager) removePublisher(ns string) {
	sm.nslock.Lock()
	defer sm.nslock.Unlock()

	delete(sm.namespaces, ns)
}

func (sm *SessionManager) getPublisher(ns string) *RelayHandler {
	sm.nslock.RLock()
	defer sm.nslock.RUnlock()

	return sm.namespaces[ns]
}

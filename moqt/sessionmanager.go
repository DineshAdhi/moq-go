package moqt

import (
	"moq-go/moqt/wire"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/rand"
)

type SessionManager struct {
	sessionslock sync.RWMutex
	nslock       sync.RWMutex
	cachelock    sync.RWMutex
	sessions     map[string]*MOQTSession
	namespaces   map[string]*MOQTSession
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{}
	sm.sessionslock = sync.RWMutex{}
	sm.nslock = sync.RWMutex{}
	sm.namespaces = map[string]*MOQTSession{}
	sm.sessions = map[string]*MOQTSession{}

	rand.Seed(uint64(time.Now().UnixNano()))

	return sm
}

func (sm *SessionManager) addSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	sm.sessions[session.id] = session

	log.Info().Msgf("[%s][New MOQT Session]", session.id)
}

func (sm *SessionManager) removeSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	id := session.id
	delete(sm.sessions, id)
}

func (sm *SessionManager) addPublisher(ns string, pub *MOQTSession) {
	sm.nslock.Lock()
	sm.namespaces[ns] = pub
	sm.nslock.Unlock()

	sm.sessionslock.RLock()

	for _, peer := range sm.sessions {

		if peer.RemoteRole == wire.ROLE_RELAY && peer.id != pub.id {
			announce := wire.Announce{
				TrackNameSpace: ns,
			}

			peer.CS.WriteControlMessage(&announce)
		}
	}

	sm.sessionslock.RUnlock()
}

func (sm *SessionManager) removePublisher(ns string) {
	sm.nslock.Lock()
	defer sm.nslock.Unlock()

	delete(sm.namespaces, ns)
}

func (sm *SessionManager) getPublisher(ns string) *MOQTSession {
	sm.nslock.RLock()
	defer sm.nslock.RUnlock()

	return sm.namespaces[ns]
}

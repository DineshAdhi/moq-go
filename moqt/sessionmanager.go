package moqt

import (
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type SessionManager struct {
	sessionslock sync.RWMutex
	nslock       sync.RWMutex
	sessions     map[string]*MOQTSession
	namespaces   map[string]*ControlHandler
}

func NewSessionManager() SessionManager {
	sm := SessionManager{}
	sm.sessionslock = sync.RWMutex{}
	sm.nslock = sync.RWMutex{}
	sm.namespaces = map[string]*ControlHandler{}
	sm.sessions = map[string]*MOQTSession{}

	rand.Seed(uint64(time.Now().UnixNano()))

	return sm
}

func (sm *SessionManager) addSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	sm.sessions[session.id] = session
}

func (sm *SessionManager) removeSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	id := session.id
	delete(sm.sessions, id)
}

func (sm *SessionManager) addNameSpace(ns string, ch *ControlHandler) {
	sm.nslock.Lock()
	defer sm.nslock.Unlock()

	sm.namespaces[ns] = ch
}

func (sm *SessionManager) removeNameSpace(ns string) {
	sm.nslock.Lock()
	defer sm.nslock.Unlock()

	delete(sm.namespaces, ns)
}

func (sm *SessionManager) getControlHandler(ns string) *ControlHandler {
	sm.nslock.RLock()
	defer sm.nslock.RUnlock()

	return sm.namespaces[ns]
}

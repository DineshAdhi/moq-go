package moqt

import (
	"fmt"
	"sync"
)

type SessionManager struct {
	rwlock   sync.RWMutex
	sessions map[string]*MOQTSession
}

func (sm *SessionManager) addSession(session *MOQTSession) {
	sm.rwlock.Lock()
	defer sm.rwlock.Unlock()

	sm.sessions[session.id] = session
}

func (sm *SessionManager) removeSession(session *MOQTSession) {
	sm.rwlock.Lock()
	defer sm.rwlock.Unlock()

	id := session.id
	delete(sm.sessions, id)
}

func (sm *SessionManager) getSessionWithNamespace(ns string) (*MOQTSession, error) {
	sm.rwlock.RLock()
	defer sm.rwlock.RUnlock()

	for _, session := range sm.sessions {
		for namespace := range session.namespaces {

			if namespace == ns {
				return session, nil
			}
		}
	}

	return nil, fmt.Errorf("no such session registered with the namespace - %s", ns)
}

func (sm *SessionManager) forwardSubscribeOk(msg *SubsribeOkMessage) {
	sm.rwlock.RLock()
	defer sm.rwlock.RUnlock()

	for _, session := range sm.sessions {
		if session.role == ROLE_PUBLISHER {
			continue
		}

		session.forwardSubscribeOk(msg)
	}
}

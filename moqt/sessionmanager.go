package moqt

import (
	"moq-go/logger"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type SessionManager struct {
	sessionslock sync.RWMutex
	nslock       sync.RWMutex
	cachelock    sync.RWMutex
	sessions     map[string]*MOQTSession
	namespaces   map[string]*MOQTSession
	ObjectStream map[string]*ObjectStream
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{}
	sm.sessionslock = sync.RWMutex{}
	sm.nslock = sync.RWMutex{}
	sm.namespaces = map[string]*MOQTSession{}
	sm.sessions = map[string]*MOQTSession{}
	sm.ObjectStream = map[string]*ObjectStream{}

	rand.Seed(uint64(time.Now().UnixNano()))

	return sm
}

func (sm *SessionManager) addSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	sm.sessions[session.id] = session

	logger.InfoLog("[%s][New MOQT Session]", session.id)
}

func (sm *SessionManager) removeSession(session *MOQTSession) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	id := session.id
	delete(sm.sessions, id)
}

func (sm *SessionManager) addPublisher(ns string, s *MOQTSession) {
	sm.nslock.Lock()
	defer sm.nslock.Unlock()

	sm.namespaces[ns] = s
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

func (sm *SessionManager) ForwardSubscribeOk(streamid string, okmsg SubscribeOkMessage) {
	sm.sessionslock.RLock()
	defer sm.sessionslock.RUnlock()

	for _, session := range sm.sessions {
		if _, ok := session.DownStreamSubMap[streamid]; ok {
			session.SendSubcribeOk(streamid, okmsg)
		}
	}
}

func (sm *SessionManager) NotifyObjectStream(os *ObjectStream) {
	sm.cachelock.RLock()
	defer sm.cachelock.RUnlock()

	streamid := os.streamid

	for _, session := range sm.sessions {
		if _, ok := session.DownStreamSubOkMap[streamid]; ok {
			session.SubscribeToStream(os)
		}
	}
}

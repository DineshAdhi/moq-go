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
	CacheData    map[string]*CacheData
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{}
	sm.sessionslock = sync.RWMutex{}
	sm.nslock = sync.RWMutex{}
	sm.namespaces = map[string]*MOQTSession{}
	sm.sessions = map[string]*MOQTSession{}
	sm.CacheData = map[string]*CacheData{}

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

func (sm *SessionManager) notifyIncomingStreams(cacheKey string) {
	sm.sessionslock.Lock()
	defer sm.sessionslock.Unlock()

	cd := sm.getCacheData(cacheKey)

	for _, session := range sm.sessions {
		if _, ok := session.DownStreamSubIDMap[cacheKey]; ok {
			go session.sendSubOkMsg(cd)
			session.incomingStreams <- cd
		}
	}
}

func (sm *SessionManager) getCacheData(cachekey string) *CacheData {
	sm.cachelock.RLock()
	defer sm.cachelock.RUnlock()

	return sm.CacheData[cachekey]
}

func (sm *SessionManager) addCacheData(cachekey string, cd *CacheData) {
	sm.cachelock.Lock()
	defer sm.cachelock.Unlock()

	logger.InfoLog("[New Cache Details][%s]", cachekey)

	sm.CacheData[cachekey] = cd
}

func (sm *SessionManager) removeCacheData(cachekey string) {
	sm.cachelock.Lock()
	defer sm.cachelock.Unlock()

	delete(sm.CacheData, cachekey)
}

package moqt

import (
	"math/rand"
	"moq-go/logger"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go"
	"golang.org/x/net/context"
)

var DEFAULT_SERVER_SETUP = ServerSetup{SelectedVersion: DRAFT_04, Params: Parameters{
	ROLE_PARAM: &IntParameter{ROLE_PARAM, ROLE_PUBSUB},
}}

var sm *SessionManager = NewSessionManager()

type MOQTConnection interface {
	AcceptStream(context context.Context) (quic.Stream, error)
	AcceptUniStream(context context.Context) (quic.ReceiveStream, error)
	CloseWithError(quic.ApplicationErrorCode, string) error
	OpenUniStreamSync(ctx context.Context) (quic.SendStream, error)
}

type MOQTSession struct {
	Conn               MOQTConnection
	controlStream      quic.Stream
	ctx                context.Context
	ishandshakedone    bool
	id                 string
	role               uint64
	cancelFunc         func()
	DownStreamSubIDMap map[string]uint64 // Map[K - CacheKey, V - SubID] - For Subscribers DownStream ID Tracking, Will be useful to Forward SubOK from Publisher
	UpStreamSubIDMap   map[uint64]string
	notifiedCacheMap   map[uint64]string
	incomingStreams    chan *CacheData
}

func CreateMOQSession(conn MOQTConnection, role uint64) *MOQTSession {
	session := &MOQTSession{}
	session.Conn = conn
	session.ctx, session.cancelFunc = context.WithCancel(context.Background())
	session.id = strings.Split(uuid.New().String(), "-")[0]
	session.role = role
	session.DownStreamSubIDMap = map[string]uint64{}
	session.UpStreamSubIDMap = map[uint64]string{}
	session.notifiedCacheMap = map[uint64]string{}
	session.incomingStreams = make(chan *CacheData, 1)

	sm.addSession(session)

	return session
}

func (s *MOQTSession) Close(code uint64, msg string) {
	s.Conn.CloseWithError(quic.ApplicationErrorCode(code), msg)
	s.cancelFunc()

	sm.removeSession(s)

	logger.ErrorLog("[%s][Closing MOQT Session][Code - %d]%s", s.id, code, msg)
}

func (s *MOQTSession) WriteControlMessage(msg MOQTMessage) {

	if s.controlStream == nil {
		logger.ErrorLog("[%s][Error Writing Control Message][CS is nil][HS - %d]", s.id, s.ishandshakedone)
		return
	}

	_, err := s.controlStream.Write(msg.GetBytes())

	if err != nil {
		logger.ErrorLog("[%s][Error Writing to Control][%s]", s.id, err)
	}

	logger.DebugLog("[%s][Dipsatching CONTROL]%s", s.id, msg.String())
}

func (s *MOQTSession) WriteStream(stream quic.SendStream, msg MOQTMessage) int {
	logger.DebugLog("[%s][Dipsatching STREAM]%s", s.id, msg.String())
	n, err := stream.Write(msg.GetBytes())

	if err != nil {
		logger.ErrorLog("[%s][Error Writing to Stream][%s]", s.id, err)
		return 0
	}

	return n
}

func (s *MOQTSession) Serve() {
	go s.handleControlStream()
	go s.handleObjectStreams()
}

func (s *MOQTSession) sendSubscribe(submsg SubscribeMessage) {

	submsg.SubscribeID = uint64(rand.Uint32())
	cacheKey := submsg.getCacheKey()

	cd := &CacheData{}
	cd.cachekey = cacheKey
	cd.trackalias = submsg.TrackAlias
	cd.trackname = submsg.TrackName
	cd.tracknamespace = submsg.TrackNamespace
	cd.buffer = []byte{}
	cd.lock = sync.RWMutex{}

	s.UpStreamSubIDMap[submsg.SubscribeID] = cd.cachekey // Temporary will get deleted after SubOK notification
	s.notifiedCacheMap[submsg.SubscribeID] = cd.cachekey // Notified cache will stay until the subscription cancel is called

	sm.addCacheData(cacheKey, cd)

	s.WriteControlMessage(&submsg)
}

func (s *MOQTSession) sendSubOkMsg(cd *CacheData) {
	cacheKey := cd.cachekey

	subid, ok := s.DownStreamSubIDMap[cacheKey]

	if !ok {
		logger.ErrorLog("[%s][Downstream Sub ID not found for Cache key][Cache Key - %s]", s.id, cacheKey)
		return
	}

	okmsg := GetSubOKMessage(subid)
	s.WriteControlMessage(&okmsg)
}

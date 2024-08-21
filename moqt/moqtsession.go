package moqt

import (
	"math/rand"
	"moq-go/logger"
	"strings"

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
	OpenUniStream() (quic.SendStream, error)
}

type MOQTSession struct {
	Conn               MOQTConnection
	controlStream      quic.Stream
	ctx                context.Context
	ishandshakedone    bool
	id                 string
	role               uint64
	cancelFunc         func()
	DownStreamSubMap   map[string]SubscribeMessage // Map[K - streamid, V - SubID] - For Subscribers DownStream ID ObjectStreaming, Will be useful to Forward SubOK from Publisher
	DownStreamSubOkMap map[string]uint64
	UpStreamSubMap     map[uint64]SubscribeMessage // Map [K - SubID, V - streamid] - To keep track of the Upstream SubID with corresponding streamid
	UpstreamSubOkMap   map[uint64]string           // Map for the SUBIDs which received SubOK
	ObjectStreamMap    map[string]*ObjectStream
	SubscribedMap      map[string]bool
	ObjectChannel      chan *ObjectDelivery
}

func CreateMOQSession(conn MOQTConnection, role uint64) *MOQTSession {
	session := &MOQTSession{}
	session.Conn = conn
	session.ctx, session.cancelFunc = context.WithCancel(context.Background())
	session.id = strings.Split(uuid.New().String(), "-")[0]
	session.role = role
	session.DownStreamSubMap = map[string]SubscribeMessage{}
	session.DownStreamSubOkMap = map[string]uint64{}
	session.UpStreamSubMap = map[uint64]SubscribeMessage{}
	session.UpstreamSubOkMap = map[uint64]string{}
	session.ObjectStreamMap = map[string]*ObjectStream{}
	session.ObjectChannel = make(chan *ObjectDelivery, 1)
	session.SubscribedMap = map[string]bool{}

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
		logger.ErrorLog("[%s][Error Writing Control Message][CS is nil][HS - %+v]", s.id, s.ishandshakedone)
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

	go s.handleSubscribedChan()
}

func (s *MOQTSession) sendSubscribe(submsg SubscribeMessage) {

	subid := uint64(rand.Uint32())
	submsg.SubscribeID = subid

	s.UpStreamSubMap[subid] = submsg // Temporary map will get deleted after SubOK notification

	s.WriteControlMessage(&submsg)
}

func (s *MOQTSession) SendSubcribeOk(streamid string, okmsg SubscribeOkMessage) {

	if submsg, ok := s.DownStreamSubMap[streamid]; ok {
		okmsg.SubscribeID = submsg.SubscribeID
		s.WriteControlMessage(&okmsg)
		s.DownStreamSubOkMap[streamid] = submsg.SubscribeID
		return
	}

	logger.ErrorLog("[%s][Error Dispatching Subscrike OK][Unable to find Subscribe for Cache Key - %s]", s.id, streamid)
}

func (s *MOQTSession) GetObjectStream(subid uint64) *ObjectStream {

	streamid, ok := s.UpstreamSubOkMap[subid]

	if !ok {
		return nil
	}

	if cd, ok := s.ObjectStreamMap[streamid]; ok {
		return cd
	}

	submsg, ok := s.UpStreamSubMap[subid]

	if !ok {
		return nil
	}

	streamid = submsg.getstreamid()

	os := NewObjectStream(streamid)
	s.ObjectStreamMap[streamid] = os

	// Notify all sessions about new Cache Data
	sm.NotifyObjectStream(os)

	return os
}

func (s *MOQTSession) SubscribeToStream(os *ObjectStream) {

	if _, ok := s.SubscribedMap[os.streamid]; ok {
		logger.DebugLog("[%s][Already Subscribed to Stream][Cache Key - %s]", s.id, os.streamid)
		return
	}

	logger.DebugLog("[%s][Subscribed to stream][%s]", os.streamid)

	os.AddSubscriber(s)

	s.SubscribedMap[os.streamid] = true
}

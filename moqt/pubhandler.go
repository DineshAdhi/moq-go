package moqt

import (
	"moq-go/moqt/wire"
	"sync"
)

type PubHandler struct {
	session          *MOQTSession
	AnnounceList     map[string]*StreamsMap
	lock             sync.RWMutex
	SubscribeChannel chan *PubStream
}

func CreateNewPubHandler(session *MOQTSession) *PubHandler {
	handler := PubHandler{
		session:          session,
		AnnounceList:     map[string]*StreamsMap{},
		lock:             sync.RWMutex{},
		SubscribeChannel: make(chan *PubStream, 10),
	}

	return &handler
}

func (pub *PubHandler) Announce(ns string) {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.session.CS.WriteControlMessage(&wire.Announce{
		TrackNameSpace: ns,
	})

	pub.AnnounceList[ns] = NewStreamsMap(pub.session)
}

func (pub *PubHandler) SubscribeOk(ps *PubStream) {

	okmsg := &wire.SubscribeOk{
		SubscribeID:   ps.SubId,
		Expires:       1024,
		ContentExists: 0,
	}

	pub.session.CS.WriteControlMessage(okmsg)

	pub.lock.Lock()
	defer pub.lock.Unlock()

	sm := pub.AnnounceList[ps.Namespace]

	if sm != nil {
		sm.AddStream(ps.SubId, ps)
		pub.session.Slogger.Info().Msgf("[Created New Stream][SubID - %x][StreamID - %s", ps.SubId, ps.StreamId)
	} else {
		pub.session.Slogger.Error().Msgf("[Error Adding Stream][Namespace not found][%s]", ps.Namespace)
	}
}

func (pub *PubHandler) HandleSubscribe(msg *wire.Subscribe) {
	pub.session.Slogger.Info().Msgf(msg.String())

	ns := msg.TrackNameSpace
	streamid := msg.GetStreamID()
	subId := msg.SubscribeID
	alias := msg.TrackAlias
	name := msg.TrackName

	pub.lock.RLock()
	defer pub.lock.RUnlock()

	if _, ok := pub.AnnounceList[ns]; ok {
		ps := NewPubStream(pub.session, ns, name, streamid, subId, alias)
		pub.SubscribeChannel <- ps
	} else {
		pub.session.Slogger.Error().Msgf("[Received Invalid Subscription][No such Namespace][%s]", ns)
	}
}

func (pub *PubHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	pub.session.Slogger.Info().Msgf(msg.String())
}

func (pub *PubHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

}

func (pub *PubHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {

}

func (pub *PubHandler) ProcessTracks() {

}

func (pub *PubHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {
	pub.session.Close(wire.MOQERR_PROTOCOL_VIOLATION, "Protocol Violation. I am a Pub, dont send SubOk to me.")
}

func (pub *PubHandler) HandleAnnounce(msg *wire.Announce) {
	pub.session.Close(wire.MOQERR_PROTOCOL_VIOLATION, "Protocol Violation. I am a Pub, dont send Announce to me.")
}

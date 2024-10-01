package moqt

import (
	"moq-go/moqt/wire"
	"sync"

	"github.com/quic-go/quic-go"
)

type PubHandler struct {
	*MOQTSession
	AnnouncedStreams map[string]*StreamsMap[*PubStream]
	NamespaceLock    sync.RWMutex
	SubscribeChan    chan PubStream
}

func NewPubHandler(session *MOQTSession) *PubHandler {
	return &PubHandler{
		MOQTSession:      session,
		AnnouncedStreams: map[string]*StreamsMap[*PubStream]{},
		NamespaceLock:    sync.RWMutex{},
		SubscribeChan:    make(chan PubStream),
	}
}

func (pub *PubHandler) SendAnnounce(namespace string) {
	announce := wire.Announce{
		TrackNameSpace: namespace,
	}

	pub.CS.WriteControlMessage(&announce)

	pub.NamespaceLock.Lock()
	defer pub.NamespaceLock.Unlock()

	smap := NewStreamsMap[*PubStream](pub.MOQTSession)
	pub.AnnouncedStreams[namespace] = &smap
}

func (pub *PubHandler) HandleAnnounce(msg *wire.Announce) {
	pub.Conn.CloseWithError(quic.ApplicationErrorCode(wire.MOQERR_PROTOCOL_VIOLATION), "I am pub. Dont send Announce to me")
}

func (pub *PubHandler) HandleSubscribe(msg *wire.Subscribe) {
	ns := msg.TrackNameSpace

	if smap, ok := pub.AnnouncedStreams[ns]; ok {

		pubstream := NewPubStream(pub.MOQTSession, msg.GetStreamID(), msg.SubscribeID, msg.TrackNameSpace, msg.TrackName, msg.TrackAlias)
		smap.AddStream(msg.SubscribeID, pubstream)

		pub.SubscribeChan <- *pubstream

	} else {
		pub.Slogger.Error().Msgf("[Received Subscribe for Unknown Namespace]")
	}
}

func (pub *PubHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {

}

func (pub *PubHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	pub.Slogger.Info().Msgf(msg.String())
}

func (pub *PubHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {
	pub.Slogger.Info().Msgf(msg.String())
}

func (pub *PubHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {

}

// SubHandler -
func (pub *PubHandler) DoHandle() {

}

func (pub *PubHandler) HandleClose() {
	close(pub.SubscribeChan)
}

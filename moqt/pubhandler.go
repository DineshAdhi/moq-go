package moqt

import "moq-go/moqt/wire"

type PubHandler struct {
	*MOQTSession
	OutgoingStreams StreamsMap
}

func CreateNewPubHandler(session *MOQTSession) *PubHandler {
	handler := PubHandler{
		MOQTSession:     session,
		OutgoingStreams: NewStreamsMap(session),
	}

	return &handler
}

func (pub *PubHandler) HandleAnnounce(msg *wire.Announce) {
	pub.Close(wire.MOQERR_PROTOCOL_VIOLATION, "Protocol Violation. I am a Pub, dont send Announce to me.")
}

func (pub *PubHandler) HandleSubscribe(msg *wire.Subscribe) {
	pub.Slogger.Info().Msgf(msg.String())
}

func (pub *PubHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {

}

func (pub *PubHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	pub.Slogger.Info().Msgf(msg.String())
}

func (pub *PubHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

}

func (pub *PubHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {

}

func (pub *PubHandler) ProcessTracks() {

}

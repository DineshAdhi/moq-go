package moqt

import "moq-go/moqt/wire"

type SubHandler struct {
	*MOQTSession
}

func CreateNewSubHandler(session *MOQTSession) *SubHandler {
	handler := SubHandler{
		MOQTSession: session,
	}

	return &handler
}

func (pub *SubHandler) HandleAnnounce(msg *wire.Announce) {
	pub.Close(wire.MOQERR_PROTOCOL_VIOLATION, "Protocol Violation. I am a Pub, dont send Announce to me.")
}

func (pub *SubHandler) HandleSubscribe(msg *wire.Subscribe) {
	pub.Slogger.Info().Msgf(msg.String())
}

func (pub *SubHandler) HandleSubscribeOk(msg *wire.SubscribeOk) {

}

func (pub *SubHandler) HandleAnnounceOk(msg *wire.AnnounceOk) {
	pub.Slogger.Info().Msgf(msg.String())
}

func (pub *SubHandler) HandleUnsubscribe(msg *wire.Unsubcribe) {

}

func (pub *SubHandler) HandleSubscribeDone(msg *wire.SubscribeDone) {

}

func (pub *SubHandler) HandleUniStreams() {

}

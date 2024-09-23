package moqt

import (
	"fmt"
	"moq-go/moqt/wire"
)

type Handler interface {
	HandleAnnounce(msg *wire.Announce)
	HandleSubscribe(msg *wire.Subscribe)
	HandleSubscribeOk(msg *wire.SubscribeOk)
	HandleAnnounceOk(msg *wire.AnnounceOk)
	HandleUnsubscribe(msg *wire.Unsubcribe)
	HandleSubscribeDone(msg *wire.SubscribeDone)
	ProcessTracks()
}

func CreateNewHandler(role uint64, session *MOQTSession) (Handler, error) {

	switch role {
	case wire.ROLE_RELAY:
		return CreateNewRelayHandler(session), nil
	case wire.ROLE_PUBLISHER:
		return CreateNewPubHandler(session), nil
	default:
		return nil, fmt.Errorf("[Local Role not Supported][%X]", role)
	}
}

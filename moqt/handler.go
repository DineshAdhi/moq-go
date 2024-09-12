package moqt

import (
	"fmt"
	"moq-go/moqt/wire"
)

type Handler interface {
	HandleAnnounce(msg *wire.AnnounceMessage)
	HandleSubscribe(msg *wire.SubscribeMessage)
	HandleSubscribeOk(msg *wire.SubscribeOkMessage)
	HandleAnnounceOk(msg *wire.AnnounceOkMessage)
}

func CreateNewHandler(role uint64, session *MOQTSession) (Handler, error) {

	switch role {
	case wire.ROLE_RELAY:
		return &RelayHandler{session}, nil
	default:
		return nil, fmt.Errorf("[Local Role not Supported][%X]", role)
	}
}

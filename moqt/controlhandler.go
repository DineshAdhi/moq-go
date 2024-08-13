package moqt

import (
	"fmt"
	"math/rand/v2"
	"moq-go/logger"

	"github.com/quic-go/quic-go/quicvarint"
)

type ControlHandler struct {
	session         *MOQTSession
	reader          quicvarint.Reader
	ishandshakedone bool
	frtable         map[uint64]*ControlHandler
	idmap           map[uint64]uint64
}

func NewControlHandler(session *MOQTSession) *ControlHandler {
	ch := &ControlHandler{}
	ch.session = session
	ch.ishandshakedone = false
	ch.frtable = map[uint64]*ControlHandler{}
	ch.idmap = map[uint64]uint64{}

	return ch
}

func (ch *ControlHandler) Close(code uint64, msg string) {
	ch.session.Close(code, msg)
}

func (ch *ControlHandler) forwardSubscribe(msg SubscribeMessage, peercontrol *ControlHandler) uint64 {
	upstreamid := uint64(rand.Uint32())
	ch.frtable[upstreamid] = peercontrol

	msg.SubscribeID = upstreamid

	ch.session.WriteControlMessage(&msg)

	return upstreamid
}

func (ch *ControlHandler) forwardSubscribeOk(msg SubsribeOkMessage) {
	upstreamid := msg.id

	if downstreamid, v := ch.idmap[upstreamid]; v {
		msg.id = downstreamid
		ch.session.WriteControlMessage(&msg)
		delete(ch.idmap, upstreamid)
	}

}

func (ch *ControlHandler) Run() {

	ch.reader = quicvarint.NewReader(ch.session.controlStream)

	for {
		moqtMessage, err := ParseMOQTMessage(ch.reader)

		if err != nil {
			ch.Close(MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Error Parsing Control Message][%s]", err))
			return
		}

		logger.DebugLog("[%s][MOQT Message]%s", ch.session.id, moqtMessage.String())

		if ch.ishandshakedone {
			ch.handleControlMessage(moqtMessage)
		} else {
			ch.handleSetupMessage(moqtMessage)
		}
	}
}

func (ch *ControlHandler) handleSetupMessage(msg MOQTMessage) {

	switch msg.Type() {
	case CLIENT_SETUP:
		clientSetup := msg.(*ClientSetup)

		if !clientSetup.CheckDraftSupport() {
			ch.session.Close(MOQERR_INTERNAL_ERROR, "CLIENT SETUP ERROR : PROTOCOL DRAFT NOT SUPPORTED")
			return
		}

		if role := clientSetup.Params.GetParameter(ROLE_PARAM); role != nil {
			ch.session.id += "_" + GetRoleString(role)
		}

		ch.session.WriteControlMessage(&DEFAULT_SERVER_SETUP)
		ch.ishandshakedone = true

	default:
		logger.ErrorLog("[Received Unknown Setup Message][Type - %s][%X]", GetMoqMessageString(msg.Type()), msg.Type())
	}
}

func (ch *ControlHandler) handleControlMessage(msg MOQTMessage) {

	switch msg.Type() {
	case ANNOUNCE:

		announceMsg := msg.(*AnnounceMessage)
		ns := announceMsg.tracknamespace

		sm.addNameSpace(ns, ch)

		okmsg := AnnounceOkMessage{}
		okmsg.tracknamespace = ns

		ch.session.WriteControlMessage(&okmsg)

	case SUBSCRIBE:

		submsg := msg.(*SubscribeMessage)
		ns := submsg.TrackNamespace
		downstreamid := submsg.SubscribeID

		if peercontrol := sm.getControlHandler(ns); peercontrol != nil {
			upstreamid := peercontrol.forwardSubscribe(*submsg, ch)
			ch.idmap[upstreamid] = downstreamid
		} else {
			logger.ErrorLog("[Error Processing Subscribe][Namespace not found][%s]", ns)
		}

	case SUBSCRIBE_OK:

		subokmsg := msg.(*SubsribeOkMessage)
		upstreamid := subokmsg.id

		if peercontrol := ch.frtable[upstreamid]; peercontrol != nil {
			peercontrol.forwardSubscribeOk(*subokmsg)
			delete(ch.frtable, upstreamid)
		} else {
			logger.ErrorLog("[Error Processing SubscribeOK][Peer Control Not found]")
		}
	}
}

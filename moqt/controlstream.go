package moqt

import (
	"fmt"
	"moq-go/logger"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

func (s *MOQTSession) handleControlStream() {

	for {
		var err error
		cs, err := s.Conn.AcceptStream(s.ctx)

		if err != nil {
			s.Close(MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Error Accepting Control Stream][%s]", err))
			return
		}

		if s.controlStream == nil {
			go s.ServeControlStream(cs)
		} else {
			s.Close(MOQERR_PROTOCOL_VIOLATION, "Received Control Stream Twice")
			return
		}
	}
}

func (s *MOQTSession) ServeControlStream(cs quic.Stream) {

	s.controlStream = cs
	reader := quicvarint.NewReader(s.controlStream)

	for {
		moqtMessage, err := ParseMOQTMessage(reader)

		if err != nil {
			s.Close(MOQERR_INTERNAL_ERROR, fmt.Sprintf("[Error Parsing Control Message][%s]", err))
			return
		}

		logger.DebugLog("[%s][MOQT Message]%s", s.id, moqtMessage.String())

		if s.ishandshakedone {
			s.handleControlMessage(moqtMessage)
		} else {
			s.handleSetupMessage(moqtMessage)
		}
	}
}

func (s *MOQTSession) handleSetupMessage(msg MOQTMessage) {

	switch msg.Type() {
	case CLIENT_SETUP:
		clientSetup := msg.(*ClientSetup)

		if !clientSetup.CheckDraftSupport() {
			s.Close(MOQERR_INTERNAL_ERROR, "CLIENT SETUP ERROR : PROTOCOL DRAFT NOT SUPPORTED")
			return
		}

		if role := clientSetup.Params.GetParameter(ROLE_PARAM); role != nil {
			s.id += "_" + GetRoleString(role)
		}

		s.ishandshakedone = true
		logger.InfoLog("[%s][Handshake Success]", s.id)
		s.WriteControlMessage(&DEFAULT_SERVER_SETUP)

	default:
		logger.ErrorLog("[Received Unknown Setup Message][Type - %s][%X]", GetMoqMessageString(msg.Type()), msg.Type())
	}
}

func (s *MOQTSession) handleControlMessage(msg MOQTMessage) {

	switch msg.Type() {
	case ANNOUNCE:

		if announceMsg, ok := msg.(*AnnounceMessage); ok {
			s.handleAnnounce(announceMsg)
		}

	case SUBSCRIBE:

		if submsg, ok := msg.(*SubscribeMessage); ok {
			s.handleSubscribe(submsg)
		}

	case SUBSCRIBE_OK:

		if okmsg, ok := msg.(*SubscribeOkMessage); ok {
			s.handleSubOk(okmsg)
		}

	}
}

func (s *MOQTSession) handleAnnounce(amsg *AnnounceMessage) {

	logger.InfoLog("[%s][ANNOUNCE][%+v]", s.id, amsg)

	if s.role == ROLE_PUBSUB || s.role == ROLE_SUBSCRIBER {

		ns := amsg.tracknamespace
		sm.addPublisher(ns, s)

		okmsg := AnnounceOkMessage{}
		okmsg.tracknamespace = ns

		s.WriteControlMessage(&okmsg)
	} else {
		s.Close(MOQERR_PROTOCOL_VIOLATION, fmt.Sprintf("Received Announce at Unsupported MOQT. Remote Role - %s", GetRoleStringVarInt(s.role)))
	}
}

func (s *MOQTSession) handleSubscribe(submsg *SubscribeMessage) {

	logger.InfoLog("[%s][SUBSCRIBE][%+v]", s.id, submsg)

	subid := submsg.SubscribeID
	tracknamespace := submsg.TrackNamespace
	cachekey := submsg.getCacheKey()

	if s.role == ROLE_PUBLISHER {
		// TODO : Handle Subscribe for Publisher Handling
	} else if s.role == ROLE_PUBSUB {

		s.DownStreamSubIDMap[cachekey] = subid

		if pubslisher := sm.getPublisher(tracknamespace); pubslisher != nil {
			pubslisher.sendSubscribe(*submsg)
		} else {
			logger.ErrorLog("[Subscribe Error][No publisher with namespace - %s]", tracknamespace)
			return
		}

	} else {
		s.Close(MOQERR_PROTOCOL_VIOLATION, fmt.Sprintf("Subscribe Unsupported for server with Role : %s", GetRoleStringVarInt(s.role)))
	}
}

func (s *MOQTSession) handleSubOk(okmsg *SubscribeOkMessage) {

	if s.role == ROLE_SUBSCRIBER || s.role == ROLE_PUBSUB {

		subid := okmsg.SubscribeID

		if cachekey, ok := s.UpStreamSubIDMap[subid]; ok {
			sm.notifyIncomingStreams(cachekey)
			delete(s.UpStreamSubIDMap, subid)
		} else {
			logger.ErrorLog("[%s][Received SubOK SubId for Unregistered Cache Key][ID - %X]", s.id, subid)
		}
	} else {
		s.Close(MOQERR_PROTOCOL_VIOLATION, fmt.Sprintf("Receive SubOk on a Unsupported Role : %s", GetRoleStringVarInt(s.role)))
	}
}
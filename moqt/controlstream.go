package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
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
			s.role = role.Value().(uint64)
			s.slogger = log.With().Str("ID", s.id).Str("Role", GetRoleStringVarInt(s.role)).Logger()
		}

		s.ishandshakedone = true
		s.slogger.Info().Msgf("[%s][Handshake Success]", s.id)
		s.WriteControlMessage(&DEFAULT_SERVER_SETUP)

	default:
		log.Error().Msgf("[Received Unknown Setup Message][Type - %s][%X]", GetMoqMessageString(msg.Type()), msg.Type())
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

	log.Info().Msgf("[%s][ANNOUNCE][%+v]", s.id, amsg)

	if s.role == ROLE_PUBSUB || s.role == ROLE_PUBLISHER {

		ns := amsg.tracknamespace
		sm.addPublisher(ns, s)

		okmsg := AnnounceOkMessage{}
		okmsg.tracknamespace = ns

		s.WriteControlMessage(&okmsg)
	} else {
		s.Close(MOQERR_PROTOCOL_VIOLATION, fmt.Sprintf("Received Announce at Unsupported  Remote Role - %s", GetRoleStringVarInt(s.role)))
	}
}

func (s *MOQTSession) handleSubscribe(submsg *SubscribeMessage) {

	log.Info().Msgf("[%s][SUBSCRIBE][%+v]", s.id, submsg)

	tracknamespace := submsg.ObjectStreamNamespace
	streamid := submsg.getstreamid()

	if s.role == ROLE_PUBLISHER {
		// TODO : Handle Subscribe for Publisher Handling
	} else if s.role == ROLE_PUBSUB || s.role == ROLE_SUBSCRIBER {

		s.DownStreamSubMap[streamid] = *submsg

		if publisher := sm.getPublisher(tracknamespace); publisher != nil {

			if cd, ok := publisher.ObjectStreamMap[streamid]; ok { // Publisher already has the Cache Data. Ignore sending SUBSCRIBE

				s.SendSubcribeOk(streamid, GetSubOKMessage(submsg.SubscribeID))
				s.SubscribeToStream(cd)

			} else {
				publisher.sendSubscribe(*submsg)
			}

		} else {
			log.Error().Msgf("[Subscribe Error][No publisher with namespace - %s]", tracknamespace)
			return
		}

	} else {
		s.Close(MOQERR_PROTOCOL_VIOLATION, fmt.Sprintf("Subscribe Unsupported for server with Role : %s", GetRoleStringVarInt(s.role)))
	}
}

func (s *MOQTSession) handleSubOk(okmsg *SubscribeOkMessage) {

	if s.role == ROLE_PUBLISHER || s.role == ROLE_PUBSUB {

		subid := okmsg.SubscribeID
		submsg, ok := s.UpStreamSubMap[subid]
		streamid := submsg.getstreamid()

		log.Info().Msgf("[%s][SUBSCRIBEOK][%s]", s.id, okmsg)

		if !ok {
			log.Error().Msgf("[%s][Received Invalid SUBSCRIBE OK][Sub Id - %X]", s.id, subid)
			return
		}

		_, ok = s.UpstreamSubOkMap[subid]

		if ok {
			log.Error().Msgf("[%s][Received Duplicate SUBSCRIKE OK][Sub Id - %X]", s.id, subid)
			return
		}

		s.UpstreamSubOkMap[subid] = streamid
		sm.ForwardSubscribeOk(streamid, *okmsg)
	} else {
		s.Close(MOQERR_PROTOCOL_VIOLATION, fmt.Sprintf("Receive SubOk on a Unsupported Role : %s", GetRoleStringVarInt(s.role)))
	}
}

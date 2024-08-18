package moqt

import (
	"moq-go/logger"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

const OBJECT_READ_DATA_LEN = 1024

func (s *MOQTSession) handleObjectStreams() {

	for {
		objstream, err := s.Conn.AcceptUniStream(s.ctx)

		if err != nil {
			logger.ErrorLog("[%s][Error Accepting Object Stream][%s]", s.id, err)
			return
		}

		go s.ServeObjectStream(objstream)
	}
}

func (s *MOQTSession) ServeObjectStream(objstream quic.ReceiveStream) {

	reader := quicvarint.NewReader(objstream)
	msg, err := ParseMOQTMessage(reader)

	if err != nil {
		logger.ErrorLog("[%s][Object Stream][Error Reading Header]", s.id)
		return
	}

	if msg.Type() != STREAM_HEADER_GROUP {
		logger.ErrorLog("[%s][Object Stream][Received Unknown Message]", s.id)
		return
	}
}

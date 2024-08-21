package moqt

import (
	"io"
	"moq-go/logger"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

const OBJECT_READ_DATA_LEN = 1024

var DUMMY_ITR = 0

func (s *MOQTSession) handleObjectStreams() {

	for {
		unistream, err := s.Conn.AcceptUniStream(s.ctx)

		if err != nil {
			logger.ErrorLog("[%s][Error Accepting Object Stream][%s]", s.id, err)
			return
		}

		go s.ServeObjectStream(unistream)
	}
}

func (s *MOQTSession) ServeObjectStream(unistream quic.ReceiveStream) {

	reader := quicvarint.NewReader(unistream)

	header, err := ParseMOQTObjectHeader(reader)

	if err != nil {
		logger.ErrorLog("[%s][Object Stream][Error Reading Header]", s.id)
		return
	}

	object := NewMOQTObject(header)
	go object.ParseFromStream(reader)

	objectStream := s.GetObjectStream(header.GetSubID())

	if objectStream == nil {
		logger.ErrorLog("Stream not found")
		return
	}

	go objectStream.addObject(object)
}

func (s *MOQTSession) handleSubscribedChan() {

	for {
		delivery := <-s.ObjectChannel

		stream := delivery.os
		object := delivery.object

		if object != nil {
			go s.DispatchObject(stream, object)
		} else {
			logger.ErrorLog("Object not found")
		}
	}
}

func (s *MOQTSession) DispatchObject(os *ObjectStream, object *MOQTObject) {
	unistream, err := s.Conn.OpenUniStream()

	if err != nil {
		logger.ErrorLog("[%s][Error dispatching object][%s]", s.id, err)
		return
	}

	subid := s.DownStreamSubOkMap[os.streamid]
	header := object.header

	unistream.Write(header.GetBytes(subid))

	reader := object.NewReader()
	data := make([]byte, OBJECT_READ_LENGTH)

	for {
		n, err := reader.Read(data)

		if err != nil {

			if err == io.EOF {
				break
			}

			logger.ErrorLog("[%s][Error Writing Object Payload]", s.id)
			return
		}

		unistream.Write(data[:n])
	}

	unistream.Close()
}

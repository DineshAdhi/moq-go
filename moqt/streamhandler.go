package moqt

import (
	"io"

	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

const OBJECT_READ_DATA_LEN = 1024

var DUMMY_ITR = 0

func (s *MOQTSession) handleObjectStreams() {

	for {
		unistream, err := s.Conn.AcceptUniStream(s.ctx)

		if err != nil {
			log.Error().Msgf("[%s][Error Accepting Object Stream][%s]", s.id, err)
			return
		}

		go s.ServeObjectStream(unistream)
	}
}

func (s *MOQTSession) ServeObjectStream(unistream quic.ReceiveStream) {

	reader := quicvarint.NewReader(unistream)

	header, err := ParseMOQTObjectHeader(reader)

	if err != nil {
		log.Error().Msgf("[%s][Object Stream][Error Reading Header]", s.id)
		return
	}

	object := NewMOQTObject(header)
	go object.ParseFromStream(reader)

	for {
		objectStream := s.GetObjectStream(header.GetSubID())

		if objectStream == nil {
			log.Error().Msgf("Stream not found")
			<-time.After(time.Second / 2)
		} else {
			go objectStream.addObject(object)
			break
		}
	}

}

func (s *MOQTSession) handleSubscribedChan() {

	for {
		delivery := <-s.ObjectChannel

		stream := delivery.os
		obj := delivery.object

		if obj != nil {
			go s.DispatchObject(stream, obj)
		} else {
			log.Error().Msgf("Object not found")
		}
	}
}

func (s *MOQTSession) DispatchObject(os *ObjectStream, object *MOQTObject) {
	unistream, err := s.Conn.OpenUniStream()

	if err != nil {
		log.Error().Msgf("[%s][Error dispatching object][%s]", s.id, err)
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

			log.Error().Msgf("[%s][Error Writing Object Payload]", s.id)
			return
		}

		unistream.Write(data[:n])
	}

	unistream.Close()
}

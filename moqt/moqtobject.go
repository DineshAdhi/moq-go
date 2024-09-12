package moqt

import (
	"io"
	"time"

	"sync"

	"moq-go/moqt/wire"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

const (
	OBJECT_READ_LENGTH = 1024
)

type MOQTObject struct {
	objlock   sync.RWMutex
	header    wire.MOQTObjectHeader
	data      []byte
	len       int
	iseof     bool
	createdat time.Time
	streamid  string
}

func NewMOQTObject(header wire.MOQTObjectHeader) *MOQTObject {
	object := &MOQTObject{}
	object.header = header
	object.data = make([]byte, 0)
	object.len = 0
	object.objlock = sync.RWMutex{}
	object.iseof = false
	object.createdat = time.Now()
	return object
}

func (object *MOQTObject) Write(buffer []byte) {
	object.objlock.Lock()
	defer object.objlock.Unlock()

	object.data = append(object.data, buffer...)
	object.len += len(buffer)
}

func (object *MOQTObject) SetStreamID(sid string) {
	object.streamid = sid
}

func (object *MOQTObject) GetStreamID() string {
	return object.streamid
}

func (object *MOQTObject) isExpired() bool {
	now := time.Now()

	if now.Sub(object.createdat).Seconds() >= 10 {
		return true
	}

	return false
}

func (object *MOQTObject) ParseFromStream(reader quicvarint.Reader) {

	var buffer []byte = make([]byte, OBJECT_READ_LENGTH)

	for {
		n, err := reader.Read(buffer)

		if err != nil {

			if err == io.EOF {
				object.Write(buffer[:n])
				object.iseof = true
				return
			}

			log.Error().Msgf("[%s][Error Reading From MOQTObject][%s]", object.header.GetObjectKey(), err)
			return
		}

		object.Write(buffer[:n])
	}
}

func (object *MOQTObject) NewReader() MOQTObjectReader {
	reader := MOQTObjectReader{}
	reader.object = object
	reader.offset = 0

	return reader
}

type MOQTObjectReader struct {
	object *MOQTObject
	offset int
}

func (r *MOQTObjectReader) Read(buffer []byte) (int, error) {

	object := r.object

	object.objlock.RLock()
	defer object.objlock.RUnlock()

	if r.offset >= object.len {
		if object.iseof {
			return 0, io.EOF
		}

		return 0, nil
	}

	n := copy(buffer, object.data[r.offset:])
	r.offset += n

	return n, nil
}

func (r *MOQTObjectReader) Pipe(stream quic.SendStream) {

	data := make([]byte, OBJECT_READ_LENGTH)

	for {
		n, err := r.Read(data)

		if err != nil {

			if err == io.EOF {
				break
			}

			log.Error().Msgf("[Error Writing Object Payload]")
			return
		}

		stream.Write(data[:n])
	}
}

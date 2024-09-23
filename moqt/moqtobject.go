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
	OBJECT_EXPIRY_TIME = 10
)

type MOQTObject struct {
	objlock   sync.RWMutex
	Header    wire.MOQTObjectHeader
	data      []byte
	len       int
	iseof     bool
	createdat time.Time
	streamid  string
	Reader    quicvarint.Reader
}

func NewMOQTObject(header wire.MOQTObjectHeader, streamid string, reader quicvarint.Reader) *MOQTObject {
	object := &MOQTObject{}
	object.Header = header
	object.data = make([]byte, 0)
	object.len = 0
	object.objlock = sync.RWMutex{}
	object.iseof = false
	object.createdat = time.Now()
	object.streamid = streamid
	object.Reader = reader
	return object
}

func (object *MOQTObject) Write(buffer []byte) {
	object.objlock.Lock()
	defer object.objlock.Unlock()

	object.data = append(object.data, buffer...)
	object.len += len(buffer)
}

func (object *MOQTObject) GetStreamID() string {
	return object.streamid
}

func (object *MOQTObject) IsExpired() bool {
	now := time.Now()

	if now.Sub(object.createdat).Seconds() >= OBJECT_EXPIRY_TIME {
		return true
	}

	return false
}

func (object *MOQTObject) ParseFromStream(reader quicvarint.Reader) {

	var buffer [OBJECT_READ_LENGTH]byte

	for {
		n, err := reader.Read(buffer[:])

		if err != nil {

			if err == io.EOF {
				object.Write(buffer[:n])
				object.iseof = true
				return
			}

			log.Error().Msgf("[%s][Error Reading From MOQTObject][%s]", object.Header.GetObjectKey(), err)
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

func (r *MOQTObjectReader) ReadToStream(stream quic.SendStream) (int, error) {

	object := r.object

	object.objlock.RLock()
	defer object.objlock.RUnlock()

	if r.offset >= object.len {
		if object.iseof {
			return 0, io.EOF
		}

		return 0, nil
	}

	n, err := stream.Write(object.data[r.offset:])

	if err != nil {
		return n, err
	}

	r.offset += n

	if r.offset >= object.len {
		if object.iseof {
			return 0, io.EOF
		}

		return 0, nil
	}

	return n, nil
}

func (r *MOQTObjectReader) Pipe(stream quic.SendStream) {

	var err error
	var n int

	for {

		if n, err = r.ReadToStream(stream); n == 0 {
			<-time.After(time.Millisecond * 10)
		}

		if err != nil {

			if err == io.EOF {
				break
			}

			log.Error().Msgf("[Error Writing Object Payload]")
			return
		}
	}
}

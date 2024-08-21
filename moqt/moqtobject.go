package moqt

import (
	"io"
	"moq-go/logger"
	"sync"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	OBJECT_READ_LENGTH = 1024
)

type MOQTObject struct {
	objlock sync.RWMutex
	header  MOQTObjectHeader
	data    []byte
	len     int
	iseof   bool
}

func NewMOQTObject(header MOQTObjectHeader) *MOQTObject {
	object := &MOQTObject{}
	object.header = header
	object.data = make([]byte, 0)
	object.len = 0
	object.objlock = sync.RWMutex{}
	object.iseof = false
	return object
}

func (object *MOQTObject) Write(buffer []byte) {
	object.objlock.Lock()
	defer object.objlock.Unlock()

	object.data = append(object.data, buffer...)
	object.len += len(buffer)
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

			logger.ErrorLog("[%s][Error Reading From MOQTObject][%s]", object.header.GetObjectKey(), err)
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

package moqt

import (
	"io"
	"moq-go/logger"
	"sync"
)

type CacheData struct {
	tracknamespace string
	trackname      string
	trackalias     uint64
	cachekey       string
	buffer         []byte
	iseof          bool
	len            int
	lock           sync.RWMutex
}

func (cd *CacheData) Write(data []byte) {
	cd.lock.Lock()
	defer cd.lock.Unlock()

	if data == nil {
		logger.DebugLog("Data NIL")
	}

	if cd.buffer == nil {
		logger.DebugLog("buffer NIL")
	}

	cd.buffer = append(cd.buffer, data...)
}

func (cd *CacheData) SetLength(n int) {
	cd.len = n
}

func (cd *CacheData) NewReader() *CacheReader {
	return &CacheReader{
		offset: 0,
		cd:     cd,
	}
}

type CacheReader struct {
	offset int
	cd     *CacheData
}

func (cr *CacheReader) Read(data []byte) (int, error) {
	cr.cd.lock.RLock()
	defer cr.cd.lock.RUnlock()

	if cr.offset >= len(cr.cd.buffer) {
		if cr.cd.iseof {
			return 0, io.EOF
		}

		return 0, nil
	}

	n := copy(data, cr.cd.buffer[cr.offset:])
	cr.offset += n

	return n, nil
}

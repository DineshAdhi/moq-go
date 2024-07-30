package wt

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/quic-go/qpack"
	"github.com/quic-go/quic-go"
)

// DataStreamer lets the caller take over the stream. After a call to DataStream
// the HTTP server library will not do anything else with the connection.
//
// It becomes the caller's responsibility to manage and close the stream.
//
// After a call to DataStream, the original Request.Body must not be used.
type DataStreamer interface {
	DataStream() quic.Stream
}

type ResWriter struct {
	stream         quic.Stream // needed for DataStream()
	bufferedStream *bufio.Writer

	header         http.Header
	status         int // status code passed to WriteHeader
	headerWritten  bool
	dataStreamUsed bool // set when DataSteam() is called
}

func NewResWriter(stream quic.Stream) *ResWriter {
	return &ResWriter{
		header:         http.Header{},
		stream:         stream,
		bufferedStream: bufio.NewWriter(stream),
	}
}

func (w *ResWriter) Header() http.Header {
	return w.header
}

func (w *ResWriter) WriteHeader(status int) {
	if w.headerWritten {
		return
	}

	if status < 100 || status >= 200 {
		w.headerWritten = true
	}
	w.status = status

	var headers bytes.Buffer
	enc := qpack.NewEncoder(&headers)
	enc.WriteField(qpack.HeaderField{Name: ":status", Value: strconv.Itoa(status)})
	for k, v := range w.header {
		for index := range v {
			enc.WriteField(qpack.HeaderField{Name: strings.ToLower(k), Value: v[index]})
		}
	}

	headersFrame := Frame{Type: FRAME_HEADERS, Length: uint64(headers.Len()), Data: headers.Bytes()}

	log.Printf("[Writing Response][%+v]", w.header)

	w.bufferedStream.Write(headersFrame.GetBytes())

	if !w.headerWritten {
		w.Flush()
	}
}

func (w *ResWriter) Write(p []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(200)
	}
	if !bodyAllowedForStatus(w.status) {
		return 0, http.ErrBodyNotAllowed
	}

	dataFrame := Frame{Type: FRAME_DATA, Length: uint64(len(p)), Data: p}
	return w.bufferedStream.Write(dataFrame.GetBytes())
}

func (w *ResWriter) Flush() {
	w.bufferedStream.Flush()
}

func (w *ResWriter) usedDataStream() bool {
	return w.dataStreamUsed
}

func (w *ResWriter) DataStream() quic.Stream {
	w.dataStreamUsed = true
	w.Flush()
	return w.stream
}

// copied from http2/http2.go
// bodyAllowedForStatus reports whether a given response status code
// permits a body. See RFC 2616, section 4.4.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}

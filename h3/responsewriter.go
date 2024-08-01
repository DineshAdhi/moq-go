package h3

import (
	"bufio"
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/quic-go/qpack"
	"github.com/quic-go/quic-go"
)

type ResponseWriter struct {
	stream         quic.Stream
	header         http.Header
	bufferedStream *bufio.Writer
}

func NewResponseWriter(stream quic.Stream) *ResponseWriter {
	return &ResponseWriter{
		stream:         stream,
		bufferedStream: bufio.NewWriter(stream),
		header:         http.Header{},
	}
}

func (w *ResponseWriter) Header() http.Header {
	return w.header
}

func (w *ResponseWriter) WriteHeader(status int) {
	f := Frame{Type: FRAME_HEADERS}

	var headers bytes.Buffer
	encoder := qpack.NewEncoder(&headers)

	encoder.WriteField(qpack.HeaderField{Name: ":status", Value: strconv.Itoa(status)})

	for k, v := range w.header {
		for index := range v {
			encoder.WriteField(qpack.HeaderField{Name: strings.ToLower(k), Value: v[index]})
		}
	}

	f.flength = uint64(headers.Len())
	f.fpayload = headers.Bytes()

	w.bufferedStream.Write(f.GetBytes())
	w.bufferedStream.Flush()
}

func (w *ResponseWriter) Write(data []byte) (int, error) {

	df := Frame{Type: FRAME_DATA, flength: uint64(len(data)), fpayload: data}

	n, err := w.bufferedStream.Write(df.GetBytes())
	w.bufferedStream.Flush()

	return n, err
}

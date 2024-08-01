package h3

import (
	"bufio"
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
	hframe := HeaderFrame{}

	var hfs []qpack.HeaderField

	hfs = append(hfs, qpack.HeaderField{Name: ":status", Value: strconv.Itoa(status)})

	for k, v := range w.header {
		for index := range v {
			hf := qpack.HeaderField{Name: strings.ToLower(k), Value: v[index]}
			hfs = append(hfs, hf)
		}
	}

	hframe.hfs = hfs

	w.bufferedStream.Write(hframe.GetBytes())
	w.bufferedStream.Flush()
}

func (w *ResponseWriter) Write(data []byte) (int, error) {

	df := DataFrame{data: data}

	n, err := w.bufferedStream.Write(df.GetBytes())
	w.bufferedStream.Flush()

	return n, err
}

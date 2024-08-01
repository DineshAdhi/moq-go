package h3

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/quic-go/qpack"
	"github.com/quic-go/quic-go/quicvarint"
)

type HeaderFrame struct {
	hfs []qpack.HeaderField
}

func (hframe *HeaderFrame) Parse(r MessageReader) error {
	length, err := quicvarint.Read(r)

	if err != nil {
		return err
	}

	data := make([]byte, length)
	r.Read(data)

	decoder := qpack.NewDecoder(nil)
	hfs, err := decoder.DecodeFull(data)

	if err != nil {
		return nil
	}

	hframe.hfs = hfs

	return nil
}

func (hframe HeaderFrame) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, FRAME_HEADERS)

	var headerdata bytes.Buffer
	encoder := qpack.NewEncoder(&headerdata)

	for _, hf := range hframe.hfs {
		encoder.WriteField(hf)
	}

	headerbytes := headerdata.Bytes()

	data = quicvarint.Append(data, uint64(len(headerbytes)))
	data = append(data, headerbytes...)

	return data
}

func (hframe HeaderFrame) WrapRequest() (*http.Request, string, error) {

	var err error
	var path, authority, method, contentLengthStr, protocol string

	httpHeaders := http.Header{}

	for _, h := range hframe.hfs {
		switch h.Name {
		case ":path":
			path = h.Value
		case ":method":
			method = h.Value
		case ":authority":
			authority = h.Value
		case ":protocol":
			protocol = h.Value
		case "content-length":
			contentLengthStr = h.Value
		default:
			if !h.IsPseudo() {
				httpHeaders.Add(h.Name, h.Value)
			}
		}
	}

	if len(httpHeaders["Cookie"]) > 0 {
		httpHeaders.Set("Cookie", strings.Join(httpHeaders["Cookie"], "; "))
	}

	var contentLength int64
	if len(contentLengthStr) > 0 {
		contentLength, err = strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			return nil, "", err
		}
	}

	isConnect := method == http.MethodConnect

	var u *url.URL
	var requestURI string

	if isConnect {
		u, err = url.ParseRequestURI("https://" + authority + path)
		if err != nil {
			return nil, "", err
		}

		requestURI = path
	} else {
		u, err = url.ParseRequestURI(path)
		if err != nil {
			return nil, "", err
		}
		requestURI = path
	}

	return &http.Request{
		Method:        method,
		URL:           u,
		Proto:         "HTTP/3",
		ProtoMajor:    3,
		ProtoMinor:    0,
		Header:        httpHeaders,
		Body:          nil,
		ContentLength: contentLength,
		Host:          authority,
		RequestURI:    requestURI,
		TLS:           &tls.ConnectionState{},
	}, protocol, nil
}

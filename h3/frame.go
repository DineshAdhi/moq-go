package h3

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/quic-go/qpack"
	"github.com/quic-go/quic-go/quicvarint"
)

const (
	FRAME_DATA         = uint64(0x00)
	FRAME_HEADERS      = uint64(0x01)
	FRAME_CANCEL_PUSH  = uint64(0x03)
	FRAME_SETTINGS     = uint64(0x04)
	FRAME_PUSH_PROMISE = uint64(0x05)
	FRAME_GOAWAY       = uint64(0x07)
	FRAME_MAX_PUSH_ID  = uint64(0x0D)
)

func GetFrameTypeString(ftype uint64) string {
	switch ftype {
	case FRAME_DATA:
		return string("FRAME_DATA")
	case FRAME_HEADERS:
		return string("FRAME_HEADERS")
	case FRAME_CANCEL_PUSH:
		return string("FRAME_CANCEL_PUSH")
	case FRAME_SETTINGS:
		return string("FRAME_SETTINGS")
	case FRAME_PUSH_PROMISE:
		return string("FRAME_PUSH_PROMISE")
	case FRAME_GOAWAY:
		return string("FRAME_GOAWAY")
	case FRAME_MAX_PUSH_ID:
		return string("FRAME_MAX_PUSH_ID")
	default:
		return string("Unknown Header Type")
	}
}

type Frame struct {
	Type     uint64
	flength  uint64
	fpayload []byte
}

func (f *Frame) Read(r io.Reader) (interface{}, error) {
	vireader := quicvarint.NewReader(r)
	ftype, err := quicvarint.Read(vireader)

	if err != nil {
		return nil, err
	}

	flength, err := quicvarint.Read(vireader)

	if err != nil {
		return nil, err
	}

	fpayload := make([]byte, flength)
	_, err = r.Read(fpayload)

	if err != nil {
		return nil, err
	}

	f.Type = ftype
	f.flength = flength
	f.fpayload = fpayload

	if f.Type == FRAME_SETTINGS {
		sframe := SettingsFrame{}
		err := sframe.Read(f)

		if err != nil {
			return nil, err
		}

		return sframe, nil
	}

	return nil, fmt.Errorf("[Received Unknown Frame][Type - %x]", ftype)
}

func (f *Frame) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, f.Type)
	data = quicvarint.Append(data, f.flength)
	data = append(data, f.fpayload...)

	return data
}

func (f *Frame) WrapRequest() (*http.Request, string, error) {
	decoder := qpack.NewDecoder(nil)
	hfields, err := decoder.DecodeFull(f.fpayload)

	if err != nil {
		return nil, "", err
	}

	var path, authority, method, contentLengthStr, protocol string

	httpHeaders := http.Header{}

	for _, h := range hfields {
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

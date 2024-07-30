package wt

import (
	"bytes"
	"io"
	"log"
	"strconv"

	"github.com/quic-go/qpack"
	"github.com/quic-go/quic-go/quicvarint"
)

type Frame struct {
	Type      uint64
	SessionID uint64
	Length    uint64
	Data      []byte
}

const (
	FRAME_DATA                = 0x00
	FRAME_HEADERS             = 0x01
	FRAME_CANCEL_PUSH         = 0x03
	FRAME_SETTINGS            = 0x04
	FRAME_PUSH_PROMISE        = 0x05
	FRAME_GOAWAY              = 0x07
	FRAME_MAX_PUSH_ID         = 0x0D
	FRAME_WEBTRANSPORT_STREAM = 0x41
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
	case FRAME_WEBTRANSPORT_STREAM:
		return string("FRAME_WEBTRANSPORT_STREAM")
	default:
		return string("Unknown Header Type")
	}
}

func (f *Frame) parse() string {

	str := bytes.Buffer{}

	str.WriteString("Type - " + GetFrameTypeString(f.Type) + " : {")

	reader := bytes.NewReader(f.Data)

	for reader.Len() > 0 {
		key, err := quicvarint.Read(reader)

		if err != nil {
			log.Println("[Error Parsing Frame][Cannot Read Frame Data]")
			return str.String()
		}

		val, err := quicvarint.Read(reader)

		if err != nil {
			log.Println("[Error Parsing Frame][Cannot Read Frame Data]")
			return str.String()
		}

		str.WriteString(GetSettingString(key) + " [" + strconv.Itoa(int(key)) + "] : " + strconv.Itoa(int(val)) + " ")
	}

	str.WriteString("}")

	return str.String()
}

func (f *Frame) Read(r io.Reader) error {
	qr := quicvarint.NewReader(r)
	t, err := quicvarint.Read(qr)
	if err != nil {
		return err
	}
	l, err := quicvarint.Read(qr)
	if err != nil {
		return err
	}

	f.Type = t

	switch t {
	case FRAME_WEBTRANSPORT_STREAM:
		f.Length = 0
		f.SessionID = l
		f.Data = []byte{}
		return nil
	default:
		f.Length = l
		f.Data = make([]byte, l)
		_, err := r.Read(f.Data)
		return err
	}
}

func (f Frame) GetBytes() []byte {
	buf := &bytes.Buffer{}

	buf.Write(quicvarint.Append(nil, f.Type))
	if f.Type == FRAME_WEBTRANSPORT_STREAM {
		buf.Write(quicvarint.Append(nil, f.SessionID))
	} else {
		buf.Write(quicvarint.Append(nil, f.Length))
	}
	buf.Write(f.Data)

	return buf.Bytes()
}

func (f Frame) decodeHeaders() ([]qpack.HeaderField, error) {
	decoder := qpack.NewDecoder(nil)
	hfields, err := decoder.DecodeFull(f.Data)

	if err != nil {
		return nil, err
	}

	return hfields, nil
}

package wt

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"moq-go/h3"
	"net/http"

	"github.com/quic-go/quic-go"
)

type WTSession struct {
	quic.Stream
	quicConn       quic.Connection
	serverStream   quic.SendStream
	clientStream   quic.ReceiveStream
	requestStream  quic.Stream
	ResponseWriter *h3.ResponseWriter
	context        context.Context
}

var DEFAULT_SETTINGS = []h3.Setting{
	{Key: h3.ENABLE_WEBTRANSPORT, Value: 1},
	{Key: h3.SETTINGS_H3_DATAGRAM, Value: 1},
	{Key: h3.WEBTRANSPORT_MAX_SESSIONS, Value: 1},
	{Key: h3.SETTINGS_ENABLE_CONNECT_PROTOCOL, Value: 1},
	{Key: h3.H3_DATAGRAM_05, Value: 1},
}

func UpgradeWTS(quicConn quic.Connection) (*WTSession, *http.Request, error) {

	// 1. Server opens a Uni-Stream and sends its Server SettingsFrame

	serverstream, err := quicConn.OpenUniStream()

	if err != nil {
		return nil, nil, err
	}

	StreamControlHeader := StreamHeader{Type: STREAM_CONTROL}
	data, err := StreamControlHeader.GetBytes()

	if err != nil {
		return nil, nil, err
	}

	serverstream.Write(data)

	serverSettingFrame := h3.SettingsFrame{Settings: DEFAULT_SETTINGS}
	serverstream.Write(serverSettingFrame.GetBytes())

	log.Printf("[Sending Server Settings][%s]", serverSettingFrame.GetString())

	// 2. Server accepts a Uni-Stream and reads the Client SettingsFrame

	clientstream, err := quicConn.AcceptUniStream(context.TODO())
	clientreader := bufio.NewReader(clientstream)

	if err != nil {
		return nil, nil, err
	}

	StreamControlHeader = StreamHeader{}
	StreamControlHeader.Read(clientstream)

	if StreamControlHeader.Type != STREAM_CONTROL {
		return nil, nil, fmt.Errorf("[Client Control Header Type Mismatch][%x]", StreamControlHeader.Type)
	}

	ftype, frame, err := h3.ParseFrame(clientreader)

	if err != nil {
		return nil, nil, err
	}

	if ftype != h3.FRAME_SETTINGS {
		return nil, nil, fmt.Errorf("[Error Receiving Settings from client][Type Mismatch][Type - %X]", ftype)
	}

	sFrame := frame.(*h3.SettingsFrame)

	log.Printf("[Received Client Settings][%s]", sFrame.GetString())

	// 3. Server now accepts Bi-Direction Stream, read headers and respond on the same stream

	rrStream, err := quicConn.AcceptStream(context.TODO()) // Request-Response Stream
	rreader := bufio.NewReader(rrStream)

	if err != nil {
		return nil, nil, err
	}

	ftype, frame, err = h3.ParseFrame(rreader)

	if err != nil {
		return nil, nil, err
	}

	if ftype != h3.FRAME_HEADERS {
		return nil, nil, fmt.Errorf("[Error Processing WT conn][Received Wrong Headers][%X]", ftype)
	}

	headerFrame := frame.(*h3.HeaderFrame)

	req, protocol, err := headerFrame.WrapRequest()

	if err != nil {
		return nil, nil, err
	}

	if protocol != "webtransport" {
		return nil, nil, fmt.Errorf("[Protocol Mismatch]")
	}

	responseWriter := h3.NewResponseWriter(rrStream)
	responseWriter.Header().Add("Sec-Webtransport-Http3-Draft", "draft02")

	wts := &WTSession{
		quicConn:       quicConn,
		clientStream:   clientstream,
		serverStream:   serverstream,
		requestStream:  rrStream,
		ResponseWriter: responseWriter,
		context:        context.TODO(),
	}

	req.Body = wts

	return wts, req, nil
}

func (wts *WTSession) AcceptSession() {
	wts.ResponseWriter.WriteHeader(200)
}

func (wts *WTSession) AcceptStream() (quic.Stream, error) {
	stream, err := wts.quicConn.AcceptStream(context.TODO())
	reader := bufio.NewReader(stream)

	if err != nil {
		return nil, err
	}

	ftype, _, err := h3.ParseFrame(reader)

	if err != nil {
		return nil, err
	}

	if ftype != STREAM_WEBTRANSPORT_BI_STREAM {
		return nil, fmt.Errorf("[Stream Header Mismatch]")
	}

	return stream, err
}

func (wts *WTSession) AcceptUniStream() (quic.ReceiveStream, error) {
	stream, err := wts.quicConn.AcceptUniStream(context.TODO())
	reader := bufio.NewReader(stream)

	if err != nil {
		return nil, err
	}

	ftype, _, err := h3.ParseFrame(reader)

	if err != nil {
		return nil, err
	}

	if ftype != STREAM_WEBTRANSPORT_UNI_STREAM {
		return nil, fmt.Errorf("[Stream Header Mismatch]")
	}

	return stream, err
}

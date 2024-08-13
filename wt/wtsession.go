package wt

import (
	"context"
	"fmt"
	"moq-go/h3"
	"moq-go/logger"
	"net/http"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
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

	controlHeader := StreamHeader{Type: STREAM_CONTROL}
	serverstream.Write(controlHeader.GetBytes())

	serverSettingFrame := h3.SettingsFrame{Settings: DEFAULT_SETTINGS}
	serverstream.Write(serverSettingFrame.GetBytes())

	logger.DebugLog("[Sending Server Settings][%s]", serverSettingFrame.GetString())

	// 2. Server accepts a Uni-Stream and reads the Client SettingsFrame

	clientstream, err := quicConn.AcceptUniStream(context.TODO())
	clientreader := quicvarint.NewReader(clientstream)

	if err != nil {
		return nil, nil, err
	}

	controlHeader = StreamHeader{}
	controlHeader.Read(clientreader)

	if controlHeader.Type != STREAM_CONTROL {
		return nil, nil, fmt.Errorf("[Client Control Header Type Mismatch][%x]", controlHeader.Type)
	}

	ftype, frame, err := h3.ParseFrame(clientreader)

	if err != nil {
		return nil, nil, err
	}

	if ftype != h3.FRAME_SETTINGS {
		return nil, nil, fmt.Errorf("[Error Receiving Settings from client][Type Mismatch][Type - %X]", ftype)
	}

	sFrame := frame.(*h3.SettingsFrame)

	logger.DebugLog("[Received Client Settings][%s]", sFrame.GetString())

	// 3. Server now accepts Bi-Direction Stream, read headers and respond on the same stream

	rrStream, err := quicConn.AcceptStream(context.TODO()) // Request-Response Stream
	rreader := quicvarint.NewReader(rrStream)

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

	logger.DebugLog("[WTS][Header Frames][%+v]", headerFrame)

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

	// Ignoring two Unistreams pushed by HTTP3 - Need to refer the draft to handle it better
	go func() {
		wts.quicConn.AcceptUniStream(context.TODO())
		wts.quicConn.AcceptUniStream(context.TODO())
	}()

	logger.DebugLog("[Accepting WebTransport][STATUS - 200]")
	wts.ResponseWriter.WriteHeader(200)
}

func (wts *WTSession) AcceptStream(ctx context.Context) (quic.Stream, error) {
	stream, err := wts.quicConn.AcceptStream(ctx)

	if err != nil {
		return nil, err
	}

	reader := quicvarint.NewReader(stream)
	header := StreamHeader{}
	header.Read(reader)

	if err != nil {
		return nil, err
	}

	if header.Type != STREAM_WEBTRANSPORT_BI_STREAM {
		return nil, fmt.Errorf("[Stream Header Mismatch]")
	}

	return stream, err
}

func (wts *WTSession) AcceptUniStream(ctx context.Context) (quic.ReceiveStream, error) {
	stream, err := wts.quicConn.AcceptUniStream(ctx)

	if err != nil {
		return nil, err
	}

	reader := quicvarint.NewReader(stream)
	header := StreamHeader{}
	header.Read(reader)

	if err != nil {
		return nil, err
	}

	if header.Type != STREAM_WEBTRANSPORT_UNI_STREAM {
		return nil, fmt.Errorf("[Stream Header Mismatch]")
	}

	return stream, err
}

func (wts *WTSession) CloseWithError(errcode quic.ApplicationErrorCode, phrase string) error {
	return wts.quicConn.CloseWithError(errcode, phrase)
}

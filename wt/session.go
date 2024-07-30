package wt

import (
	"context"
	"encoding/binary"
	"log"
	"net/http"

	"github.com/quic-go/quic-go"
)

type ServerControlStream struct {
	quic.SendStream
}

type ClientControlStream struct {
	quic.ReceiveStream
}

func (CCS *ServerControlStream) WriteFrame(f Frame) int {
	n, _ := CCS.Write(f.GetBytes())
	return n
}

func (CCS *ServerControlStream) WriteStreamHeader(s StreamHeader) int {
	data, err := s.GetBytes()

	if err != nil {
		log.Println("CCS : Error Writing Frame", err)
		return -1
	}

	n, err := CCS.Write(data)

	return n
}

type WTSession struct {
	quic.Stream
	conn      quic.Connection
	CCS       ClientControlStream // Server Control Stream
	scs       ServerControlStream // Client Control Stream
	rrs       quic.Stream
	ResWriter *ResWriter
	context   context.Context
}

func NewWTSession(conn quic.Connection, ctx context.Context) (*WTSession, *http.Request, error) {

	// ServerControlStream - Opens a UniStream

	serverControlStream, err := conn.OpenUniStream()

	if err != nil {
		return nil, nil, err
	}

	// Sending STREAM_CONTROL & SETTINGS_FRAME

	var scs ServerControlStream = ServerControlStream{serverControlStream}

	stream_control_header := StreamHeader{Type: STREAM_CONTROL}
	n := scs.WriteStreamHeader(stream_control_header)

	if n > 0 {
		// log.Println("[SCS][Stream Control Packet][Dispatched]")
	}

	settings := DefaultSettings()
	serverSettingsFrame := settings.ToFrame()
	n = scs.WriteFrame(serverSettingsFrame)

	if n > 0 {
		log.Printf("[SCS][Server Settings Frame][Dispatched][%s]", serverSettingsFrame.parse())
	}

	// ClientControlStream - Accepts a UniStream

	clientControlStream, err := conn.AcceptUniStream(context.Background())

	if err != nil {
		return nil, nil, err
	}

	// Checks if STREAM_CONTROL HEADER && SETTINGS_FRAME ARE VALID

	var CCS ClientControlStream = ClientControlStream{clientControlStream}

	client_stream_header := StreamHeader{}
	client_stream_header.Read(CCS)

	if client_stream_header.Type != STREAM_CONTROL {
		log.Println("[CCS][Error Creating Session][Received Wrong Control Header]")
		return nil, nil, nil
	}

	settingsFrame := Frame{}
	err = settingsFrame.Read(CCS)

	log.Printf("[CCS][Client Settings Frame][Received][%s]", settingsFrame.parse())

	if err != nil || settingsFrame.Type != FRAME_SETTINGS {
		log.Printf("[CCS][Received Invalid Frame][Frame Type Invalid][Type Received - %x]", settingsFrame.Type)
		return nil, nil, err
	}

	go func() {
		for {
			var data uint8
			err := binary.Read(CCS, binary.LittleEndian, &data)

			if err != nil {
				break
			}

			log.Printf("Data - %X", data)
		}

		// for {
		// 	frame := Frame{}
		// 	frame.Read(CCS)

		// 	if err != nil {
		// 		break
		// 	}

		// 	log.Printf("FRAME - %+v", frame)
		// }
	}()

	// Request Response Stream

	rrstream, err := conn.AcceptStream(ctx)

	if err != nil {
		log.Printf("[Error Acceping Stream]")
		return nil, nil, err
	}

	headerFrame := Frame{}
	err = headerFrame.Read(rrstream)

	if err != nil {
		log.Printf("[RRS][Error Reading Frame]")
		return nil, nil, err
	}

	if headerFrame.Type != FRAME_HEADERS {
		log.Printf("[RRS][Invalid Header Frame]")
		return nil, nil, err
	}

	hfs, err := headerFrame.decodeHeaders()

	if err != nil {
		log.Printf("[RRS][Error Decoding Frames]")
		return nil, nil, err
	}

	req, protocol, err := RequestFromHeaders(hfs)

	if req.Method == "CONNECT" {
		log.Printf("[Received CONNECT][%+v]", hfs)
	}

	if err != nil {
		log.Printf("[RRS][Error Forming Requests]")
		return nil, nil, err
	}

	if protocol != "webtransport" {
		log.Printf("[RRS][Invalid Protocol][%s]", protocol)
		return nil, nil, err
	}

	rw := NewResWriter(rrstream)
	rw.Header().Add("Sec-Webtransport-Http3-Draft", "draft02")

	session := &WTSession{
		conn:      conn,
		CCS:       CCS,
		scs:       scs,
		rrs:       rrstream,
		context:   ctx,
		ResWriter: rw,
	}

	req.Body = session

	return session, req, nil
}

func (wts *WTSession) AcceptSession() {
	rw := wts.ResWriter
	rw.WriteHeader(http.StatusOK)
	rw.Flush()
}

func (wts *WTSession) AcceptStream() (quic.Stream, error) {
	stream, err := wts.conn.AcceptStream(wts.context)

	if err != nil {
		return nil, err
	}

	frame := Frame{}
	err = frame.Read(stream)

	return stream, err
}

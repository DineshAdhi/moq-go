package wire

import (
	"fmt"
	"io"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

type TrackStream struct {
	SubID      uint64
	StreamID   string
	TrackAlias uint64
	SendOrder  uint64
	ObjectsArr []*Object
	ObjectLock sync.Mutex
	ObjectCond sync.Cond
	IsEOF      bool
	wg         *sync.WaitGroup
	reader     quicvarint.Reader
}

func NewTrackStream(subid uint64, alias uint64) *TrackStream {

	ts := &TrackStream{}
	ts.SubID = subid
	ts.IsEOF = false
	ts.ObjectLock = sync.Mutex{}
	ts.ObjectCond = *sync.NewCond(&ts.ObjectLock)
	ts.wg = &sync.WaitGroup{}

	// log.Info().Msgf("[New MOQT Stream][Stream ID - %s]", streamid)

	return ts
}

func (ts *TrackStream) Parse(subid uint64, reader quicvarint.Reader) error {

	ts.SubID = subid
	ts.IsEOF = false
	ts.ObjectLock = sync.Mutex{}
	ts.ObjectCond = *sync.NewCond(&ts.ObjectLock)
	ts.wg = &sync.WaitGroup{}
	ts.reader = reader

	var err error

	if ts.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if ts.SendOrder, err = quicvarint.Read(reader); err != nil {
		return err
	}

	return nil
}

func (ts *TrackStream) WgAdd() {
	ts.wg.Add(1)
}

func (ts *TrackStream) WgWait() {
	ts.wg.Wait()
}

func (ts *TrackStream) WgDone() {
	ts.wg.Done()
}

func (ts *TrackStream) SetStreamID(streamid string) {
	ts.StreamID = streamid
}

func (ts *TrackStream) GetStreamID() string {
	return ts.StreamID
}

func (ts *TrackStream) SetReader(reader quicvarint.Reader) {
	ts.reader = reader
}

func (ts *TrackStream) GetHeaderBytes() []byte {

	var data []byte

	data = quicvarint.Append(data, STREAM_HEADER_TRACK)
	data = quicvarint.Append(data, ts.SubID)
	data = quicvarint.Append(data, ts.TrackAlias)
	data = quicvarint.Append(data, ts.SendOrder)

	return data
}

func (ts *TrackStream) GetHeaderSubIDBytes(subid uint64) []byte {

	var data []byte

	data = quicvarint.Append(data, STREAM_HEADER_TRACK)
	data = quicvarint.Append(data, subid)
	data = quicvarint.Append(data, ts.TrackAlias)
	data = quicvarint.Append(data, ts.SendOrder)

	return data
}

func (ts *TrackStream) Pipe(index int, stream quic.SendStream) (int, error) {
	ts.ObjectCond.L.Lock()
	ts.ObjectCond.Wait()

	length := len(ts.ObjectsArr)

	var data []byte

	for index < length {
		obj := ts.ObjectsArr[index]
		data = quicvarint.Append(data, obj.GroupID)
		data = append(data, obj.GetBytes()...)
		index++
	}

	ts.ObjectCond.L.Unlock()

	if _, err := stream.Write(data); err != nil {
		return index, err
	}

	if ts.IsEOF == true {
		return index, io.EOF
	}

	return index, nil
}

func (ts *TrackStream) ReadObject() (uint64, *Object, error) {

	if ts.reader == nil {
		return 0, nil, fmt.Errorf("Reader is nil")
	}

	groupid, err := quicvarint.Read(ts.reader)

	if err == io.EOF {
		ts.Close()
		return 0, nil, err
	}

	if err != nil {
		return 0, nil, err
	}

	object := &Object{}
	err = object.Parse(ts.reader)

	if err == io.EOF {
		ts.Close()
		return 0, nil, err
	}

	if err != nil {
		return 0, nil, err
	}

	object.GroupID = groupid

	ts.ObjectCond.L.Lock()
	ts.ObjectsArr = append(ts.ObjectsArr, object)
	ts.ObjectCond.Broadcast()
	ts.ObjectCond.L.Unlock()

	return groupid, object, nil
}

func (ts *TrackStream) WriteObject(object *Object) {
	ts.ObjectCond.L.Lock()
	ts.ObjectsArr = append(ts.ObjectsArr, object)
	ts.ObjectCond.Broadcast()
	ts.ObjectCond.L.Unlock()
}

func (ts *TrackStream) Close() {
	ts.ObjectCond.L.Lock()
	ts.IsEOF = true
	ts.ObjectCond.Broadcast()
	ts.ObjectCond.L.Unlock()
}

package wire

import (
	"fmt"
	"io"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

type GroupStream struct {
	SubID      uint64
	TrackAlias uint64
	StreamID   string
	GroupID    uint64
	SendOrder  uint64
	ObjectsArr []*Object
	ObjectLock sync.RWMutex
	ObjectCond *sync.Cond
	IsEOF      bool
	wg         *sync.WaitGroup
	reader     quicvarint.Reader
}

func NewGroupStream(subid uint64, grouid uint64, alias uint64) *GroupStream {

	gs := &GroupStream{}
	gs.SubID = subid
	gs.TrackAlias = alias
	gs.SendOrder = 0
	gs.GroupID = grouid
	gs.ObjectLock = sync.RWMutex{}
	gs.IsEOF = false
	gs.wg = &sync.WaitGroup{}
	gs.ObjectsArr = make([]*Object, 0)

	return gs
}

func (gs *GroupStream) Parse(subid uint64, reader quicvarint.Reader) error {

	gs.SubID = subid
	gs.ObjectLock = sync.RWMutex{}
	gs.IsEOF = false
	gs.wg = &sync.WaitGroup{}
	gs.ObjectsArr = make([]*Object, 0)
	gs.reader = reader

	var err error

	if gs.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if gs.GroupID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if gs.SendOrder, err = quicvarint.Read(reader); err != nil {
		return err
	}

	return nil
}

func (gs *GroupStream) SetStreamID(streamid string) {
	gs.StreamID = streamid
}

func (gs *GroupStream) GetStreamID() string {
	return gs.StreamID
}

func (gs *GroupStream) SetReader(reader quicvarint.Reader) {
	gs.reader = reader
}

func (gs *GroupStream) GetHeaderBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, STREAM_HEADER_GROUP)
	data = quicvarint.Append(data, gs.SubID)
	data = quicvarint.Append(data, gs.TrackAlias)
	data = quicvarint.Append(data, gs.GroupID)
	data = quicvarint.Append(data, gs.SendOrder)

	return data
}

func (gs *GroupStream) GetHeaderSubIDBytes(subid uint64) []byte {
	var data []byte

	data = quicvarint.Append(data, STREAM_HEADER_GROUP)
	data = quicvarint.Append(data, subid)
	data = quicvarint.Append(data, gs.TrackAlias)
	data = quicvarint.Append(data, gs.GroupID)
	data = quicvarint.Append(data, gs.SendOrder)

	return data
}

func (gs *GroupStream) Pipe(index int, stream quic.SendStream) (int, error) {

	gs.ObjectLock.RLock()

	length := len(gs.ObjectsArr)

	if index == length && gs.IsEOF == true {
		gs.ObjectLock.RUnlock()
		return index, io.EOF
	}

	var data []byte

	for index < length {
		obj := gs.ObjectsArr[index]
		data = append(data, obj.GetBytes()...)
		index++
	}

	gs.ObjectLock.RUnlock()

	if _, err := stream.Write(data); err != nil {
		return index, err
	}

	return index, nil
}

func (gs *GroupStream) ReadObject() (uint64, *Object, error) {

	if gs.reader == nil {
		return 0, nil, fmt.Errorf("Reader is nil")
	}

	object := &Object{}
	err := object.Parse(gs.reader)

	if err == io.EOF {
		gs.Close()
		return 0, nil, err
	}

	if err != nil {
		return 0, nil, err
	}

	gs.ObjectLock.Lock()
	gs.ObjectsArr = append(gs.ObjectsArr, object)
	gs.ObjectLock.Unlock()

	return gs.GroupID, object, nil
}

func (gs *GroupStream) WriteObject(object *Object) {
	gs.ObjectLock.Lock()
	gs.ObjectsArr = append(gs.ObjectsArr, object)
	gs.ObjectLock.Unlock()
}

func (gs *GroupStream) Close() {
	gs.ObjectLock.Lock()
	gs.IsEOF = true
	gs.ObjectLock.Unlock()
}

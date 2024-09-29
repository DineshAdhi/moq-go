package wire

import (
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
	reader     quicvarint.Reader
	ObjectsArr []*Object
	ObjectLock sync.Mutex
	ObjectCond sync.Cond
	IsEOF      bool
}

func NewGroupStream(subid uint64, streamid string, reader quicvarint.Reader) (*GroupStream, error) {

	var err error

	gs := &GroupStream{}
	gs.SubID = subid
	gs.StreamID = streamid
	gs.reader = reader
	gs.ObjectLock = sync.Mutex{}
	gs.ObjectCond = *sync.NewCond(&gs.ObjectLock)
	gs.IsEOF = false

	if gs.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return nil, err
	}

	if gs.GroupID, err = quicvarint.Read(reader); err != nil {
		return nil, err
	}

	if gs.SendOrder, err = quicvarint.Read(reader); err != nil {
		return nil, err
	}

	// log.Info().Msgf("[New MOQT Group Stream][Stream ID - %s]", streamid)

	return gs, nil
}

func (gs *GroupStream) GetStreamID() string {
	return gs.StreamID
}

func (gs *GroupStream) GetHeaderBytes() []byte {
	var data []byte

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

	gs.ObjectCond.L.Lock()
	gs.ObjectCond.Wait()

	length := len(gs.ObjectsArr)

	gs.ObjectCond.L.Unlock()

	for index < length {
		obj := gs.ObjectsArr[index]
		_, err := stream.Write(obj.GetBytes())

		if err != nil {
			return index, err
		}

		index++
	}

	if gs.IsEOF == true {
		return index, io.EOF
	}

	return index, nil
}

func (gs *GroupStream) ReadObject() (uint64, *Object, error) {

	object := &Object{}
	err := object.Parse(gs.reader)

	if err == io.EOF {
		gs.ObjectCond.L.Lock()
		gs.IsEOF = true
		gs.ObjectCond.L.Unlock()

		gs.ObjectCond.Broadcast()

		return 0, nil, err
	}

	if err != nil {
		return 0, nil, err
	}

	gs.ObjectCond.L.Lock()
	gs.ObjectsArr = append(gs.ObjectsArr, object)
	gs.ObjectCond.L.Unlock()

	gs.ObjectCond.Broadcast()

	return gs.GroupID, object, nil
}

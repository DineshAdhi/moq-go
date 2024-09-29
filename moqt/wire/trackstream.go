package wire

import (
	"io"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/rs/zerolog/log"
)

type TrackStream struct {
	SubID      uint64
	StreamID   string
	TrackAlias uint64
	SendOrder  uint64
	reader     quicvarint.Reader
	ObjectsArr []*Object
	ObjectLock sync.Mutex
	ObjectCond sync.Cond
	IsEOF      bool
}

func NewTrackStream(subid uint64, streamid string, reader quicvarint.Reader) (*TrackStream, error) {

	var err error

	ts := &TrackStream{}
	ts.SubID = subid
	ts.StreamID = streamid
	ts.IsEOF = false
	ts.ObjectLock = sync.Mutex{}
	ts.ObjectCond = *sync.NewCond(&ts.ObjectLock)

	if ts.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return nil, err
	}

	if ts.SendOrder, err = quicvarint.Read(reader); err != nil {
		return nil, err
	}

	log.Info().Msgf("[New MOQT Stream][Stream ID - %s]", streamid)

	return ts, nil
}

func (ts *TrackStream) GetStreamID() string {
	return ts.StreamID
}

func (ts *TrackStream) GetHeaderBytes() []byte {

	var data []byte

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

	ts.ObjectCond.L.Unlock()

	for index <= length {
		obj := ts.ObjectsArr[index]
		_, err := stream.Write(obj.GetBytes())

		if err != nil {
			return index, err
		}

		index++
	}

	return index, nil
}

func (ts *TrackStream) ReadObject() (uint64, *Object, error) {

	groupid, err := quicvarint.Read(ts.reader)

	if err == io.EOF {
		ts.ObjectCond.L.Lock()
		ts.IsEOF = true
		ts.ObjectCond.L.Unlock()

		ts.ObjectCond.Broadcast()

		return 0, nil, err
	}

	if err != nil {
		return 0, nil, err
	}

	object := &Object{}
	err = object.Parse(ts.reader)

	if err == io.EOF {
		ts.ObjectCond.L.Lock()
		ts.IsEOF = true
		ts.ObjectCond.L.Unlock()

		ts.ObjectCond.Broadcast()

		return 0, nil, err
	}

	if err != nil {
		return 0, nil, err
	}

	ts.ObjectCond.L.Lock()
	ts.ObjectsArr = append(ts.ObjectsArr, object)
	ts.ObjectCond.L.Unlock()

	ts.ObjectCond.Broadcast()

	return groupid, object, nil
}

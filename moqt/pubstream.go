package moqt

import (
	"io"
	"moq-go/moqt/wire"
)

type PubStream struct {
	*MOQTSession
	StreamID  string
	SubID     uint64
	Namespace string
	TrackName string
	Alias     uint64
}

func NewPubStream(session *MOQTSession, streamid string, subid uint64, ns string, trackname string, alias uint64) *PubStream {
	return &PubStream{
		MOQTSession: session,
		StreamID:    streamid,
		SubID:       subid,
		Namespace:   ns,
		TrackName:   trackname,
		Alias:       alias,
	}
}

func (pub *PubStream) Accept() {
	okmsg := wire.SubscribeOk{
		SubscribeID:   pub.SubID,
		Expires:       1024,
		ContentExists: 0,
	}

	pub.CS.WriteControlMessage(&okmsg)
}

func (pub *PubStream) NewGroup(id uint64) (wire.MOQTStream, error) {
	stream := wire.NewGroupStream(pub.SubID, id, pub.Alias)
	return pub.NewStream(stream)
}

func (pub *PubStream) NewTrack() (wire.MOQTStream, error) {
	stream := wire.NewTrackStream(pub.SubID, pub.Alias)
	return pub.NewStream(stream)
}

func (pub *PubStream) NewStream(stream wire.MOQTStream) (wire.MOQTStream, error) {

	stream.WgAdd()

	unistream, err := pub.Conn.OpenUniStream()

	if err != nil {
		return stream, err
	}

	unistream.Write(stream.GetHeaderBytes())

	go func() {
		itr := 0

		stream.WgDone()

		for {
			itr, err = stream.Pipe(itr, unistream)

			if err == io.EOF {
				break
			}

			if err != nil {
				pub.Slogger.Error().Msgf("[Error Piping Objects][%s]", err)
			}
		}

		unistream.Close()
	}()

	stream.WgWait()

	return stream, nil
}

func (pub PubStream) GetStreamID() string {
	return pub.StreamID
}

func (pub PubStream) GetSubID() uint64 {
	return pub.SubID
}

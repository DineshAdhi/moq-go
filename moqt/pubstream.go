package moqt

import (
	"io"

	"github.com/DineshAdhi/moq-go/moqt/wire"
)

type PubStream struct {
	session      *MOQTSession
	StreamID     string
	SubID        uint64
	Namespace    string
	TrackName    string
	Alias        uint64
	GroupCounter uint64
}

func NewPubStream(session *MOQTSession, streamid string, subid uint64, ns string, trackname string, alias uint64) *PubStream {
	return &PubStream{
		session:      session,
		StreamID:     streamid,
		SubID:        subid,
		Namespace:    ns,
		TrackName:    trackname,
		Alias:        alias,
		GroupCounter: 0,
	}
}

func (pub *PubStream) Accept() {
	okmsg := wire.SubscribeOk{
		SubscribeID:   pub.SubID,
		Expires:       1024,
		ContentExists: 0,
	}

	pub.session.CS.WriteControlMessage(&okmsg)
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

	unistream, err := pub.session.Conn.OpenUniStream()

	if err != nil {
		return stream, err
	}

	_, err = unistream.Write(stream.GetHeaderBytes())

	if err != nil {
		return stream, err
	}

	go func() {
		itr := 0

		for {
			itr, err = stream.Pipe(itr, unistream)

			if err == io.EOF {
				break
			}

			if err != nil {
				pub.session.Slogger.Error().Msgf("[Error Piping Objects][%s]", err)
			}
		}

		unistream.Close()
	}()

	return stream, nil
}

func (pub PubStream) GetStreamID() string {
	return pub.StreamID
}

func (pub PubStream) GetSubID() uint64 {
	return pub.SubID
}

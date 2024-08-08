package moqt

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

// SUBSCRIBE Message {
// 	Subscribe ID (i),
// 	Track Alias (i),
// 	Track Namespace (b),
// 	Track Name (b),
// 	StartGroup (Location),
// 	StartObject (Location),
// 	EndGroup (Location),
// 	EndObject (Location),
// 	Number of Parameters (i),
// 	Track Request Parameters (..) ...
//   }

type SubscribeMessage struct {
	SubscribeID    uint64
	TrackAlias     uint64
	TrackNamespace string
	TrackName      string
	StartGroup     Location
	StartObject    Location
	EndGroup       Location
	EndObject      Location
	Params         Parameters
}

func (s *SubscribeMessage) Parse(reader quicvarint.Reader) (err error) {

	if s.SubscribeID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.TrackAlias, err = quicvarint.Read(reader); err != nil {
		return err
	}

	if s.TrackNamespace, err = ParseVarIntString(reader); err != nil {
		return err
	}

	if s.TrackName, err = ParseVarIntString(reader); err != nil {
		return err
	}

	s.StartGroup = Location{}
	err = s.StartGroup.Parse(reader)

	if err != nil {
		return err
	}

	s.StartObject = Location{}
	err = s.StartObject.Parse(reader)

	if err != nil {
		return err
	}

	s.EndGroup = Location{}
	err = s.EndGroup.Parse(reader)

	if err != nil {
		return err
	}

	s.EndObject = Location{}
	err = s.EndObject.Parse(reader)

	if err != nil {
		return err
	}

	params := Parameters{}
	err = params.Parse(reader)

	if err != nil {
		return err
	}

	s.Params = params

	return nil
}

func (s SubscribeMessage) GetBytes() []byte {
	var data []byte
	return data
}

func (s SubscribeMessage) String() string {
	str := fmt.Sprintf("[%s][ID - %x][Track Name - %s][Name Space - %s][%s]", GetMoqMessageString(SUBSCRIBE), s.SubscribeID, s.TrackName, s.TrackNamespace, s.Params.String())
	return str
}

func (s SubscribeMessage) Type() uint64 {
	return SUBSCRIBE
}

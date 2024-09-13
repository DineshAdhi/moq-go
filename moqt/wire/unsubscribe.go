package wire

import (
	"fmt"

	"github.com/quic-go/quic-go/quicvarint"
)

type Unsubcribe struct {
	SubscriptionID uint64
}

func (msg *Unsubcribe) Parse(reader quicvarint.Reader) (err error) {

	if msg.SubscriptionID, err = quicvarint.Read(reader); err != nil {
		return err
	}

	return nil
}

func (msg *Unsubcribe) GetBytes() []byte {
	var data []byte
	data = quicvarint.Append(data, UNSUBSCRIBE)
	data = quicvarint.Append(data, msg.SubscriptionID)

	return data
}

func (msg *Unsubcribe) String() string {
	return fmt.Sprintf("[UNSUBSCRIBE][ID - %d]", msg.SubscriptionID)
}

func (msg *Unsubcribe) Type() uint64 {
	return UNSUBSCRIBE
}

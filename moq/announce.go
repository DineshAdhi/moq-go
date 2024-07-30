package moq

import (
	"bufio"
	"io"
	"log"

	"github.com/quic-go/quic-go/quicvarint"
)

type Announce struct {
	TrackNamespace string
	Params         []byte
}

func (a *Announce) Read(r io.Reader) error {
	reader := bufio.NewReader(r)
	// mtype, err := quicvarint.Read(reader)

	// if err != nil {
	// 	log.Println("[Error Parsing Annouce]")
	// 	return err
	// }

	// if mtype != ANNOUNCE {
	// 	hexvalue := fmt.Sprint("%x", mtype)
	// 	log.Printf("[Announce][Invalid Mtype][%s]", hexvalue)
	// 	return err
	// }

	for {
		value, err := quicvarint.Read(reader)

		if err != nil {
			break
		}

		log.Printf("DATA - %x", value)
	}

	// var size uint
	// binary.Read(r, binary.BigEndian, &size)

	// namebytes := make([]byte, size)
	// binary.Read(r, binary.BigEndian, &namebytes)

	// namespace := string(namebytes)

	// paramslen, err := quicvarint.Read(reader)

	// if err != nil {
	// 	log.Println("[Error Parsing Annouce Params]")
	// 	return err
	// }

	// log.Printf("size - %d NAMESPACE - %s", size, namespace)

	// a.TrackNamespace = namespace

	// for range paramslen {

	// }

	return nil
}

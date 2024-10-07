package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"time"

	"github.com/DineshAdhi/moq-go/moqt"
	"github.com/DineshAdhi/moq-go/moqt/api"
	"github.com/DineshAdhi/moq-go/moqt/wire"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
)

const RELAY = "localhost:4443"

var ALPN = []string{"moq-00"}
var qconfig = &quic.Config{}

var dialerOptions = moqt.DialerOptions{
	QuicConfig: qconfig,
	ALPNs:      ALPN,
}

const MD5DataType = 0x69

type MD5Data struct {
	datalen uint64
	data    []byte
	hash    [16]byte
}

func NewMD5Data() MD5Data {
	datalen := mrand.Uint32() % 1024
	data := make([]byte, datalen)
	rand.Read(data)

	hash := md5.Sum(data)

	return MD5Data{
		datalen: uint64(datalen),
		data:    data,
		hash:    hash,
	}
}

func (d *MD5Data) GetBytes() []byte {
	var data []byte

	data = quicvarint.Append(data, MD5DataType)
	data = quicvarint.Append(data, uint64(d.datalen))
	data = append(data, d.data...)
	data = append(data, d.hash[:]...)

	return data
}

func (d *MD5Data) Parse(data []byte) error {
	reader := bytes.NewReader(data)
	datatype, err := quicvarint.Read(reader)

	if err != nil {
		return err
	}

	if datatype != MD5DataType {
		return fmt.Errorf("Type Malformed - %X", datatype)
	}

	d.datalen, err = quicvarint.Read(reader)

	if err != nil {
		return err
	}

	d.data = make([]byte, d.datalen)

	_, err = io.ReadFull(reader, d.data)

	if err != nil {
		return err
	}

	_, err = io.ReadFull(reader, d.hash[:])

	if err != nil {
		return err
	}

	return nil
}

func (d *MD5Data) VerifyHash() bool {
	hash := md5.Sum(d.data)

	if hash == d.hash {
		return true
	}

	return false
}

func handlePS(ps moqt.PubStream) {
	ps.Accept()

	for gid := uint64(0); gid <= 100; gid++ {

		gs, err := ps.NewGroup(gid)

		if err != nil {
			panic(err)
		}

		for oid := uint64(0); oid < 1; oid++ {

			d := NewMD5Data()

			obj := &wire.Object{
				ID:      oid,
				Payload: d.GetBytes(),
			}

			gs.WriteObject(obj)
		}

		gs.Close()

		<-time.After(time.Millisecond * 1)
	}
}

func handleSS(ss moqt.SubStream) {

	for os := range ss.StreamsChan {
		handleOs(os.(*wire.GroupStream))
	}
}

func handleOs(os *wire.GroupStream) {

	for {
		gid, obj, err := os.ReadObject()

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		d := &MD5Data{}
		err = d.Parse(obj.Payload)

		if err != nil {
			panic(err)
		}

		if !d.VerifyHash() {
			log.Printf("Hash Verification Failed : GroupId - %d ObjectId - %d", gid, obj.ID)
			panic(d)
		} else {
			log.Printf("Hash Verified GroupId - %d ObjectId - %d", gid, obj.ID)
		}
	}

}

func NewPub() {

	pub := api.NewMOQPub(dialerOptions, RELAY)

	pub.OnSubscribe(func(ps moqt.PubStream) {
		go handlePS(ps)
	})

	handler, err := pub.Connect()

	if err != nil {
		panic(err)
	}

	handler.SendAnnounce("md5test")
}

func NewSub() {

	sub := api.NewMOQSub(dialerOptions, RELAY)

	sub.OnStream(func(ss moqt.SubStream) {
		go handleSS(ss)
	})

	handler, err := sub.Connect()

	if err != nil {
		panic(err)
	}

	handler.Subscribe("md5test", "testdata", 0)
}

func main() {
	NewPub()
	NewSub()

	select {}
}

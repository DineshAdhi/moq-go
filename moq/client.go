package moq

import (
	"io"
	"log"

	"github.com/quic-go/quic-go/quicvarint"
)

const (
	DRAFT_00 = 0xff000000
	DRAFT_01 = 0xff000001
	DRAFT_02 = 0xff000002
	DRAFT_03 = 0xff000003
)

const (
	ROLE_PARAM = 0x00
	PATH_PARAM = 0x01
)

const (
	Publisher  = 0x01
	SUBSCRIBER = 0x02
	PUBSUB     = 0x03
)

func GetRoleString(rtype uint64) string {
	switch rtype {
	case Publisher:
		return "Publisher"
	case SUBSCRIBER:
		return "SUBSCRIBER"
	case PUBSUB:
		return "PUBSUB"
	default:
		return "UNKNOWN ROLE"
	}
}

func GetParamterTypeString(ptype uint64) string {
	switch ptype {
	case ROLE_PARAM:
		return "ROLE_PARAM"
	case PATH_PARAM:
		return "PATH_PARAM"
	default:
		return "UNKNOWN_PARAM"
	}
}

type SetupParameter struct {
	ptype  uint64
	pvalue uint64
}

type ClientSetup struct {
	SupportedVersions []uint64
	Params            []SetupParameter
}

func (cs *ClientSetup) Read(r io.Reader) error {

	bytesReader := quicvarint.NewReader(r)

	mtype, err := quicvarint.Read(bytesReader) // MessageType - MOQT Message

	if err != nil {
		return err
	}

	if mtype != CLIENT_SETUP {
		log.Printf("[Received Invalid Mtype][%s]", GetMoqMessageString(mtype))
		return nil
	}

	nversions, err := quicvarint.Read(bytesReader)

	if err != nil {
		log.Printf("[Client Setup][Error While Parsing][%s]", err)
		return err
	}

	for range nversions {
		version, err := quicvarint.Read(bytesReader)

		if err != nil {
			log.Printf("[Client Setup][Error While Parsing][%s]", err)
			return err
		}

		cs.SupportedVersions = append(cs.SupportedVersions, version)
	}

	nparams, err := quicvarint.Read(bytesReader)

	if err != nil {
		log.Printf("[Error Reading nParams][%s]", err)
		return err
	}

	for range nparams {
		ptype, err := quicvarint.Read(bytesReader)

		if err != nil {
			log.Printf("[Client Setup][Error While Parsing][%s]", err)
			return err
		}

		_, err = quicvarint.Read(bytesReader)

		if err != nil {
			log.Printf("[Client Setup][Error While Parsing][%s]", err)
			return err
		}

		pvalue, err := quicvarint.Read(bytesReader) // Setup Parameter is always encoded as uint64. So, ignoring len here

		if err != nil {
			log.Printf("[Client Setup][Error While Parsing][%s]", err)
			return err
		}

		param := SetupParameter{}
		param.ptype = ptype
		param.pvalue = pvalue

		cs.Params = append(cs.Params, param)
	}

	for _, version := range cs.SupportedVersions {
		if version == DRAFT_03 {
			return nil
		}
	}

	log.Printf("[CLIENT SETUP HAS UNSUPPORTED DRAFT VERSION]")

	return err
}

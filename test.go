package main

import (
	"log"
	"moq-go/moq"
	"moq-go/wt"
	"net/http"
)

const addr = "0.0.0.0:4433"

func main() {

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		session := req.Body.(*wt.WTSession)
		session.AcceptSession()

		bistream, err := session.AcceptStream()

		if err != nil {
			log.Printf("[Error Accepting Stream]", err)
		}

		clientsetup := &moq.ClientSetup{}
		err = clientsetup.Read(bistream)

		if err != nil {
			log.Printf("[Error Reading Client Setup]", err)
		}

		log.Printf("[Got CLIENT SETUP][%+v]", clientsetup)

		serverSetup := moq.DefaultServerSetup()
		bistream.Write(serverSetup.GetBytes())

		log.Printf("[Sent SERVER SETUP][%+v]", serverSetup)
	})

	http.HandleFunc("/counter", func(rw http.ResponseWriter, req *http.Request) {
		session := req.Body.(*wt.WTSession)
		session.ResWriter.WriteHeader(200)
		session.ResWriter.Flush()

		log.Printf("Counter Triggered")
	})

	http.HandleFunc("/fingerprint", func(rw http.ResponseWriter, req *http.Request) {
		session := req.Body.(*wt.WTSession)
		session.ResWriter.WriteHeader(200)
		session.ResWriter.Flush()

		log.Printf("FP Triggered")
	})

	wts := wt.WTServer{
		ListenAddr: addr,
		CertPath:   "./certs/localhost.crt",
		KeyPath:    "./certs/localhost.key",
		ALPNs:      []string{"h3"},
	}

	err := wts.Run()

	log.Print(err)
}

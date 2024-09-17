clean:
	rm bin/relay

relay : examples/relay/relay.go clean
	go build -o bin/relay examples/relay/relay.go

run : relay
	bin/relay -certpath=./examples/certs/localhost.crt -keypath=./examples/certs/localhost.key

run_prod : relay
	env MOQT_CERT_PATH=/etc/letsencrypt/live/dineshadhi.com/fullchain.pem MOQT_KEY_PATH=/etc/letsencrypt/live/dineshadhi.com/privkey.pem bin/relay

cert:
	sh examples/certs/cert.sh

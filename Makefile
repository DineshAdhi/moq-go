clean:
	rm -f bin/relay bin/pub bin/sub

relaysource : examples/relay/relay.go clean
	go build -o bin/relay examples/relay/relay.go


pubsource : examples/pub/pub.go clean
	go build -o bin/pub examples/pub/pub.go

subsource : examples/sub/sub.go clean
	go build -o bin/sub examples/sub/sub.go

relay : relaysource
	bin/relay -certpath=./examples/certs/localhost.crt -keypath=./examples/certs/localhost.key -debug

pub : pubsource
	bin/pub -certpath=./examples/certs/localhost.crt -keypath=./examples/certs/localhost.key -debug

sub : subsource
	bin/sub -certpath=./examples/certs/localhost.crt -keypath=./examples/certs/localhost.key -debug

relay_prod : relaysource
	env MOQT_CERT_PATH=/etc/letsencrypt/live/in.dineshadhi.com/fullchain.pem MOQT_KEY_PATH=/etc/letsencrypt/live/in.dineshadhi.com/privkey.pem bin/relay

cert:
	sh examples/certs/cert.sh

go get filippo.io/mkcert
go run filippo.io/mkcert -ecdsa -install

CERTPATH=./examples/certs/

go run filippo.io/mkcert -ecdsa -cert-file ${CERTPATH}/localhost.crt -key-file ${CERTPATH}/localhost.key localhost 127.0.0.1 ::1

if [ ! -d "../certs" ]; then
  mkdir certs
fi

sh cert.sh ../certs

go run relay.go -debug=true

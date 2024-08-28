
# MOQT - GO

Simple Implentation of Media Over QUIC Transport (MOQT) in Go, in compliant with the [DRAFT04](https://datatracker.ietf.org/doc/draft-ietf-moq-transport/04/)

The Publisher and Consumer part are still under development. This repository currently has a working implementation of Relay.

This MOQT library currently supports WebTransport Protocol. Support for QUIC based MOQT is underway.

# Setup

- Keep the self-signed certificates under `certs` directory inside `examples` folder
- Run the following command

	`go run relay.go`

- You can enable the debug logs by appending `-debug=true` to this command
- Visit the [quic.video](https://quic.video/publish?server=localhost:4443) site and start publishing.
- Copy paste the shared URL to view the stream in another tab.

---


# MOQT - GO

Simple Implentation of Media Over QUIC Transport (MOQT) in Go, in compliant with the [DRAFT04](https://datatracker.ietf.org/doc/draft-ietf-moq-transport/04/)

The Publisher and Consumer part are still under development. This repository currently has a working implementation of Relay.

This MOQT library currently supports WebTransport and QUIC Protocols.

| Module    | Support |
| -------- | ------- |
| Relay  | :white_check_mark:    |
| Publisher | 	:x:     |
| Subscriber    | 	:x:   |


# Setup

- Run the following command in `examples/relay` folder

	`sh relay.sh`

- You can configure the options like shown below

  `sh relay.sh -port=<PORT> -keypath=<PATH> -certpath=<PATH> -debug`

- This will setup your self-signed certificates, add it to your Keystore and start the MOQT Relay.
- Visit the [quic.video](https://quic.video/publish?server=localhost:4443) site and start publishing. Make sure the browser supports WebTranport.
- Copy paste the shared URL to view the stream in another tab.

---

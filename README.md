
# MOQT - GO

Simple Implentation of Media Over QUIC Transport (MOQT) in Go, in compliant with the [DRAFT04](https://dataObjectStreamer.ietf.org/doc/draft-ietf-moq-transport/04/)

This MOQT library currently supports WebTransport and QUIC Protocols.

| Module    | Support |
| -------- | ------- |
| Relay  | :white_check_mark:    |
| Publisher | 	:white_check_mark:     |
| Subscriber    | 	:white_check_mark:   |


# Setup

- Configure Self-Signed Certificates by calling ```make cert```
- Implementations of Relay, Publisher and Subscriber are configured in the `examples` folder.
- You can run them using the make commands

	- `make relay`
    - `make sub`
    - `make pub`

---

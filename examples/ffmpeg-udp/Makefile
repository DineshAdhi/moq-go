pubsource : pub/pub.go
	go build -o bin/pub pub/pub.go

pub : pubsource
	bin/pub

subsource : sub/sub.go
	go build -o bin/sub sub/sub.go

sub : subsource
	bin/sub

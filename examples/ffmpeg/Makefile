pubsource : pub/pub.go
	go build -o bin/pub pub/pub.go

pub : pubsource
	ffmpeg -re -f lavfi -i testsrc=size=1280x720:rate=30 -g 30 -keyint_min 30 -vcodec libx264 -preset ultrafast -f mpegts pipe:1 | bin/pub

subsource : sub/sub.go
	go build -o bin/sub sub/sub.go

sub : subsource
	bin/sub  | ffplay -

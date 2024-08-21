package moqt

import (
	"fmt"
	"moq-go/logger"
	"strconv"
	"sync"
)

type ObjectDelivery struct {
	os     *ObjectStream
	object *MOQTObject
}

type ObjectStream struct {
	streamid    string
	trackalias  uint64
	objects     map[string]*MOQTObject
	maplock     *sync.RWMutex
	subscribers []*MOQTSession
	sublock     *sync.RWMutex
}

func NewObjectStream(streamid string, trackalias uint64) *ObjectStream {
	os := &ObjectStream{}
	os.streamid = streamid
	os.trackalias = trackalias
	os.maplock = &sync.RWMutex{}
	os.objects = map[string]*MOQTObject{}
	os.sublock = &sync.RWMutex{}
	os.subscribers = []*MOQTSession{}

	logger.InfoLog("[New Object Stream][%s]", streamid)

	return os
}

func (stream *ObjectStream) AddSubscriber(s *MOQTSession) {
	stream.sublock.Lock()
	defer stream.sublock.Unlock()

	stream.subscribers = append(stream.subscribers, s)

	objectkey := fmt.Sprintf("%s_%s", strconv.FormatUint(stream.trackalias, 10), strconv.FormatUint(0, 10))
	object := stream.getObject(objectkey)

	if object != nil {
		s.ObjectChannel <- &ObjectDelivery{stream, object}
	}
}

func (stream *ObjectStream) NotifySubscribers(object *MOQTObject, objkey string) {

	stream.sublock.RLock()
	defer stream.sublock.RUnlock()

	for _, s := range stream.subscribers {
		s.ObjectChannel <- &ObjectDelivery{stream, object}
	}
}

func (stream *ObjectStream) addObject(object *MOQTObject) {
	stream.maplock.Lock()
	defer stream.maplock.Unlock()

	objkey := object.header.GetObjectKey()
	stream.objects[objkey] = object

	go stream.NotifySubscribers(object, objkey)
}

func (stream *ObjectStream) getObject(objkey string) *MOQTObject {
	stream.maplock.RLock()
	defer stream.maplock.RUnlock()

	return stream.objects[objkey]
}

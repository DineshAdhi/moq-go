package moqt

import (
	"moq-go/logger"
	"sync"
)

type ObjectDelivery struct {
	os  *ObjectStream
	obj *MOQTObject
}

type ObjectStream struct {
	streamid    string
	objects     map[string]*MOQTObject
	maplock     *sync.RWMutex
	subscribers []*MOQTSession
	sublock     *sync.RWMutex
}

func NewObjectStream(streamid string) *ObjectStream {
	os := &ObjectStream{}
	os.streamid = streamid
	os.maplock = &sync.RWMutex{}
	os.objects = map[string]*MOQTObject{}
	os.subscribers = make([]*MOQTSession, 0)
	os.sublock = &sync.RWMutex{}

	logger.InfoLog("[New Object Stream][%s]", streamid)

	return os
}

func (stream *ObjectStream) AddSubscriber(s *MOQTSession) {
	stream.sublock.Lock()
	defer stream.sublock.Unlock()

	stream.subscribers = append(stream.subscribers, s)
}

func (stream *ObjectStream) NotifySubscribers(obj *MOQTObject) {

	for _, s := range stream.subscribers {
		s.ObjectChannel <- &ObjectDelivery{stream, obj}
	}
}

func (stream *ObjectStream) addObject(obj *MOQTObject) {
	// stream.maplock.Lock()
	// defer stream.maplock.Unlock()

	// objkey := obj.header.GetObjectKey()
	// stream.objects[objkey] = obj

	go stream.NotifySubscribers(obj)
}

func (stream *ObjectStream) getObject(objkey string) *MOQTObject {
	stream.maplock.RLock()
	defer stream.maplock.RUnlock()

	return stream.objects[objkey]
}

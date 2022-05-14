package xevent

import (
	"errors"
	"reflect"
	"sync"
)

// event stream framework

type Event struct {
	once     sync.Once
	sendLock chan struct{}
	active   subscribers

	l     sync.Mutex // protect ready sub
	ready subscribers
	eType reflect.Type
}

func (e *Event) init() {
	e.sendLock = make(chan struct{}, 1)
	e.sendLock <- struct{}{}
}

func (e *Event) Subscribe(ch interface{}) (Subscription, error) {
	e.once.Do(e.init)
	chType := reflect.TypeOf(ch)
	if chType.Kind() != reflect.Chan || chType.ChanDir()&reflect.SendDir == 0 {
		return nil, errors.New("subscribe channel fail")
	}
	e.l.Lock()
	defer e.l.Unlock()
	if !e.validateEventType(chType.Elem()) {
		return nil, errors.New("subscribe event type fail")
	}
	chValue := reflect.ValueOf(ch)
	subCase := reflect.SelectCase{Dir: reflect.SelectSend, Chan: chValue}
	e.ready = append(e.ready, subCase)
	return &subscriber{e: e, ch: chValue, err: make(chan error, 1)}, nil
}

func (e *Event) Send(value interface{}) (int, error) {
	e.once.Do(e.init)
	<-e.sendLock

	rValue := reflect.ValueOf(value)
	e.l.Lock()
	e.active = append(e.active, e.ready...)
	e.ready = nil
	if !e.validateEventType(rValue.Type()) {
		e.sendLock <- struct{}{}
		e.l.Unlock()
		return 0, errors.New("send value type wrong")
	}
	e.l.Unlock()
	for i := 0; i < len(e.active); i++ {
		e.active[i].Send = rValue
	}
	var nSend int
	list := e.active
	for {
		for i := 0; i < len(list); i++ {
			if !list[i].Chan.TrySend(rValue) {
				continue
			}
			list = list.deactivate(i)
			nSend++
			i--
		}

		if len(list) == 0 {
			break
		}

		chosen, _, _ := reflect.Select(list)
		nSend++
		list = list.deactivate(chosen)
	}

	for i := 0; i < len(e.active); i++ {
		e.active[i].Send = reflect.Value{}
	}
	e.sendLock <- struct{}{}
	return nSend, nil
}

func (e *Event) validateEventType(eType reflect.Type) bool {
	if e.eType == nil {
		e.eType = eType
		return true
	}
	return e.eType == eType
}

func (e *Event) rmv(sub *subscriber) {
	ch := sub.ch.Interface()
	e.l.Lock()
	index := e.ready.find(ch)
	if index != -1 {
		e.ready = e.ready.delete(index)
		e.l.Unlock()
		return
	}
	e.l.Unlock()
	select {
	case <-e.sendLock:
		e.active = e.active.delete(e.active.find(ch))
		e.sendLock <- struct{}{}
	}
}

type subscriber struct {
	e    *Event
	ch   reflect.Value
	once sync.Once
	err  chan error
}

func (sub *subscriber) Unsubscribe() {
	sub.once.Do(func() {
		sub.e.rmv(sub)
		close(sub.err)
	})
}

func (sub *subscriber) Err() <-chan error {
	return sub.err
}

type subscribers []reflect.SelectCase

func (list subscribers) find(ch interface{}) int {
	for index, l := range list {
		if l.Chan.Interface() == ch {
			return index
		}
	}
	return -1
}

func (list subscribers) delete(index int) subscribers {
	return append(list[:index], list[index+1:]...)
}

func (list subscribers) deactivate(index int) subscribers {
	tail := len(list) - 1
	list[index], list[tail] = list[tail], list[index]
	return list[:tail]
}

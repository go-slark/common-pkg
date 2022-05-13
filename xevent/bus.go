package xevent

import (
	"errors"
	"reflect"
	"sync"
)

type Bus struct {
	once      sync.Once
	sendLock  chan struct{}
	rmvSub    chan interface{}
	activeSub subList

	l        sync.Mutex
	readySub subList
	eType    reflect.Type
}

func (b *Bus) init() {
	b.rmvSub = make(chan interface{})
	b.sendLock = make(chan struct{}, 1)
	b.sendLock <- struct{}{}
	b.activeSub = subList{{Chan: reflect.ValueOf(b.rmvSub), Dir: reflect.SelectRecv}}
}

func (b *Bus) Subscribe(ch interface{}) (Subscription, error) {
	b.once.Do(b.init)
	chType := reflect.TypeOf(ch)
	if chType.Kind() != reflect.Chan || chType.ChanDir()&reflect.SendDir == 0 {
		return nil, errors.New("subscribe channel fail")
	}
	b.l.Lock()
	defer b.l.Unlock()
	if !b.validateEventType(chType.Elem()) {
		return nil, errors.New("subscribe event type fail")
	}
	chValue := reflect.ValueOf(ch)
	subCase := reflect.SelectCase{Dir: reflect.SelectSend, Chan: chValue}
	b.readySub = append(b.readySub, subCase)
	return &busSub{bus: b, ch: chValue, err: make(chan error, 1)}, nil
}

func (b *Bus) Send(value interface{}) (int, error) {
	b.once.Do(b.init)
	//是否可以不需要这个send lock
	<-b.sendLock

	rValue := reflect.ValueOf(value)
	b.l.Lock()
	b.activeSub = append(b.activeSub, b.readySub...)
	b.readySub = nil
	if !b.validateEventType(rValue.Type()) {
		b.sendLock <- struct{}{}
		b.l.Unlock()
		return 0, errors.New("send value type wrong")
	}
	b.l.Unlock()
	//设在所有通道上的发送至
	activeSubNum := len(b.activeSub)
	for i := 1; i < activeSubNum; i++ {
		b.activeSub[i].Send = rValue
	}
	var nSend int //发送消息到了多少个通道
	//开始发送,首先不阻塞的发送一遍
	list := b.activeSub
	for {
		for i := 1; i < activeSubNum; i++ {
			if !list[i].Chan.TrySend(rValue) {
				continue
			}
			nSend++
			i--
			list = list.deactivate(i)
		}

		if len(list) == 1 {
			break
		}

		chosen, recv, _ := reflect.Select(list)
		if chosen == 0 {
			// TODO
		} else {
			nSend++
			list = list.deactivate(chosen)
		}
	}

	for i := 1; i < activeSubNum; i++ {
		b.activeSub[i].Send = reflect.Value{}
	}
	b.sendLock <- struct{}{}
	return nSend, nil
}

func (b *Bus) validateEventType(eType reflect.Type) bool {
	if b.eType == nil {
		b.eType = eType
		return true
	}
	return b.eType == eType
}

func (b *Bus) rmv(sub *busSub) {
	ch := sub.ch.Interface()
	b.l.Lock()
	index := b.readySub.find(ch)
	if index != -1 {
		b.readySub = b.readySub.delete(index)
		b.l.Unlock()
		return
	}
	b.l.Unlock()
	select {
	case b.rmvSub <- ch:
		//TODO
	case <-b.sendLock:
		b.activeSub = b.activeSub.delete(b.activeSub.find(ch))
		b.sendLock <- struct{}{}
	}
}

type busSub struct {
	bus  *Bus
	ch   reflect.Value
	once sync.Once
	err  chan error
}

func (bs *busSub) Unsubscribe() {
	//TODO
	bs.bus.rmv(bs)
}

func (bs *busSub) Err() <-chan error {
	return bs.err
}

type subList []reflect.SelectCase

func (list subList) find(ch interface{}) int {
	for index, l := range list {
		if l.Chan.Interface() == ch {
			return index
		}
	}
	return -1
}

func (list subList) delete(index int) subList {
	return append(list[:index], list[index+1:]...)
}

func (list subList) deactivate(index int) subList {
	tail := len(list) - 1
	list[index], list[tail] = list[tail], list[index]
	return list[:tail]
}

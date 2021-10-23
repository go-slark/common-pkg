package domainevent

import (
	"errors"
	"fmt"
	"github.com/smallfish-root/common-pkg/xevent"
	"reflect"
	"time"
)

func newEventRetry() *eventRetry {
	return &eventRetry{
		pubPool: make(map[string]reflect.Type),
		subPool: make(map[string]subRetry),
		delay:   60 * time.Second, //新插入数据，延迟60秒被扫描
		retries: 3,                //重试次数
	}
}

type eventRetry struct {
	pubPool map[string]reflect.Type
	subPool map[string]subRetry
	delay   time.Duration
	retries int
}

type subRetry struct {
	event    reflect.Type
	function interface{}
}

func (retry *eventRetry) bindRetryPubEvent(event xevent.Event) {
	topic := event.Topic()
	if topic == "" {
		panic("Topic Cannot be empty")
	}
	if _, ok := retry.pubPool[topic]; ok {
		panic(fmt.Sprintf("Topic:%s already exists", topic))
	}
	retry.pubPool[topic] = reflect.TypeOf(event)
}

func (retry *eventRetry) bindRetrySubEvent(event xevent.Event, fun interface{}) {
	topic := event.Topic()
	if topic == "" {
		panic("Topic Cannot be empty")
	}

	if _, ok := retry.subPool[topic]; ok {
		panic(fmt.Sprintf("Topic:%s already exists", topic))
	}
	eventType := reflect.TypeOf(event)
	eventInType, err := parseRetryCallFunc(fun)
	if err != nil {
		panic(err)
	}
	if eventType != eventInType {
		panic("The function's argument is wrong")
	}

	retry.subPool[topic] = subRetry{
		event:    eventType,
		function: fun,
	}
}

func (retry *eventRetry) pubExist(topic string) bool {
	_, ok := retry.pubPool[topic]
	return ok
}

func (retry *eventRetry) subExist(topic string) bool {
	_, ok := retry.subPool[topic]
	return ok
}

func (retry *eventRetry) setRetryPolicy(delay time.Duration, retries int) {
	retry.delay = delay
	retry.retries = retries
}

func (retry *eventRetry) retry() {
	retry.scanSub()
	retry.scanPub()
}

func (retry *eventRetry) scanSub() {
	//每次取200条
	rows := 200
	var list []*subEventObject
	err := GetEventManager().Where("retries < ? and created < ?", retry.retries, time.Now().Add(-retry.delay)).Order("sequence ASC").Limit(rows).Find(&list).Error
	if err != nil {
		return
	}

	var filterList []*subEventObject
	for _, po := range list {
		if !retry.subExist(po.Topic) {
			GetEventManager().db().Where("identity = ?", po.Identity).Delete(&subEventObject{}) //未注册重试，直接删除
			continue
		}

		po.AddRetries(1)
		err := GetEventManager().db().Table(po.TableName()).Where(po.Location()).Updates(po.GetChanges()).Error //修改重试次数
		if err != nil {
			continue
		}
		filterList = append(filterList, po)
	}

	for _, po := range filterList {
		retry.callSub(po)
	}
}

func (retry *eventRetry) callSub(po *subEventObject) {
	defer func() {
		if perr := recover(); perr != nil {
			return
		}
	}()

	subRetryObj := retry.subPool[po.Topic]
	newEvent := reflect.New(subRetryObj.event.Elem()).Interface()

	domainEvent := newEvent.(xevent.Event)
	if err := domainEvent.Unmarshal([]byte(po.Content)); err != nil {
		return
	}
	domainEvent.SetIdentity(po.Identity)

	reflect.ValueOf(subRetryObj.function).Call([]reflect.Value{reflect.ValueOf(newEvent)})
}

func (retry *eventRetry) scanPub() {
	rows := 200
	var list []*pubEventObject
	err := GetEventManager().db().Where("retries < ? and created < ?", retry.retries, time.Now().Add(-retry.delay)).Order("sequence ASC").Limit(rows).Find(&list).Error
	if err != nil {
		return
	}

	var filterList []*pubEventObject
	for _, po := range list {
		if !retry.pubExist(po.Topic) {
			GetEventManager().db().Where("identity = ?", po.Identity).Delete(&pubEventObject{}) //未注册重试，直接删除
			continue
		}

		po.AddRetries(1)
		err := GetEventManager().db().Table(po.TableName()).Where(po.Location()).Updates(po.GetChanges()).Error //修改重试次数
		if err != nil {
			continue
		}
		filterList = append(filterList, po)
	}

	for _, po := range filterList {
		retry.callPub(po)
	}
}

func (retry *eventRetry) callPub(po *pubEventObject) {
	defer func() {
		if perr := recover(); perr != nil {
			return
		}
	}()

	pubRetryObj := retry.pubPool[po.Topic]
	//构建领域事件对象
	domainEvent := reflect.New(pubRetryObj.Elem()).Interface().(xevent.Event)
	if err := domainEvent.Unmarshal([]byte(po.Content)); err != nil {
		return
	}

	domainEvent.SetIdentity(po.Identity)
	GetEventManager().push(domainEvent)
}

func parseRetryCallFunc(f interface{}) (inType reflect.Type, e error) {
	ftype := reflect.TypeOf(f)
	if ftype.Kind() != reflect.Func {
		e = errors.New("it's not a func")
		return
	}
	if ftype.NumIn() != 1 {
		e = errors.New("the function's argument is wrong")
		return
	}
	if ftype.NumOut() != 0 {
		e = errors.New("the function's argument is wrong")
		return
	}
	inType = ftype.In(0)
	if inType.Kind() != reflect.Ptr {
		e = errors.New("the function's argument is wrong")
		return
	}
	return
}

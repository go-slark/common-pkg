package xbus

import (
	"context"
	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
	"github.com/pkg/errors"
	"github.com/smallfish-root/eventhorizon"
	"github.com/smallfish-root/eventhorizon/eventbus/kafka"
	"time"
)

//event
//事件数据

type EventData = eventhorizon.EventData
type Event struct {
	EType eventhorizon.EventType
	EId   uuid.UUID //失败的需要存起来，定时重新发布 3s
	EventData
}

func RegisterEventData(events ...Event) {
	for _, event := range events {
		eventhorizon.RegisterEventData(event.EType, func() eventhorizon.EventData { return event })
	}
}

func QueryEventData(eventType eventhorizon.EventType) (eventhorizon.EventData, error) {
	return eventhorizon.CreateEventData(eventType)
}

//事件需要嵌入到具体的命令中一起执行，另外也需要支持可单端执行(定时执行失败的eid时使用)

var eventBus *kafka.EventBus

func InitEventBus(addr []string, appId string) {
	bus, err := kafka.NewEventBus(addr, appId)
	if err != nil {
		panic(errors.WithStack(err))
	}

	eventBus = bus
}

type EventMatcher = eh.EventMatcher
type EventHandler = eh.EventHandler
type EventHandlerInfo struct {
	EventHandler
	EventMatcher
}

//具体事件需要实现响应的接口函数
//HandleEvent(ctx context.Context, event eh.Event)
//HandlerType() eh.EventHandlerType

func RegisterEventHandler(eventHandlerInfos ...EventHandlerInfo) {
	ctx := context.TODO()
	for _, eventHandlerInfo := range eventHandlerInfos {
		//需要重新修改AddHandler函数 --- 删除掉kafka reader部分  add handler是消费event?
		err := eventBus.AddHandler(ctx, eventHandlerInfo.EventMatcher, eventHandlerInfo.EventHandler)
		if err != nil {
			panic(errors.WithStack(err))
		}
	}
}

//具体event需要实现Event接口函数

//event注册时重新new一个event

//func NewEvent() *Event {
//	return eventhorizon.NewEvent("", nil, time.Now())
//	return &Event{}
//}

// NewEvent 需要发布的event
func NewEvent() eh.Event {
	return eh.NewEvent("", nil, time.Now())
}

//func (e *Event) EventType() eh.EventType {
//	return e.EventType()
//}
//
//func (e *Event) Data() eh.EventData {
//	return nil
//}
//
//func (e *Event) Timestamp() time.Time {
//	return time.Now()
//}
//
//func (e *Event) AggregateType() eh.AggregateType {
//	return ""
//}
//
//func (e *Event) AggregateID() uuid.UUID {
//	return e.EId
//}
//
//func (e *Event) Version() int {
//	return 0
//}
//
//func (e *Event) Metadata() map[string]interface{} {
//	return nil
//}
//
//func (e *Event) String() string {
//	return ""
//}

func HandleEvent(event *Event) error {
	//e, _ := QueryEventData("")
	return eventBus.HandleEvent(context.TODO(), NewEvent())
}

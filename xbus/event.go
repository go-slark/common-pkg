package xbus

import (
	"context"
	"fmt"
	eh "github.com/looplab/eventhorizon"
	kafkaEvent "github.com/looplab/eventhorizon/eventbus/kafka"
	"github.com/pkg/errors"
)

var eventBus eh.EventBus

// InitEventBus 暂时支持kafka单节点
func InitEventBus(addr, appId string) {
	eb, err := kafkaEvent.NewEventBus(addr, appId)
	if err != nil {
		panic(errors.WithStack(err))
	}

	eventBus = eb
}

func WatchEventError() {
	for err := range eventBus.Errors() {
		fmt.Printf("event bus error:%+v\n", err)
	}
}

type EventSet struct {
	//eh.MatchEvents {}
	Matcher eh.EventMatcher
	//event需要实现两个相关的接口
	Handler eh.EventHandler
}

// RegisterEventHandler 订阅event 处理
func RegisterEventHandler(eventSet ...*EventSet) {
	for _, es := range eventSet {
		err := eventBus.AddHandler(context.TODO(), es.Matcher, es.Handler)
		if err != nil {
			panic(errors.WithStack(err))
		}
	}
}

func HandleEvent(ctx context.Context, event eh.Event) error {
	//event publish
	return eventBus.HandleEvent(ctx, event)
}

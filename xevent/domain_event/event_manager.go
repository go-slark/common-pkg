package domainevent

import (
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xevent"
	mq "github.com/smallfish-root/common-pkg/xkafka"
	"github.com/smallfish-root/common-pkg/xmysql"
	"gorm.io/gorm"
	"time"
)

type EventManager struct {
	*gorm.DB
	eventRetry *eventRetry
	KafkaClent *mq.KafkaClient
}

var eventManager = make(map[string]*EventManager)

type EventManagerConf struct {
	Alias     string
	OpType int
	KafkaConf *mq.KafkaConf
}

func InitEventManager(emcs []*EventManagerConf) {
	for _, emc := range emcs {
		eventManager[emc.Alias].DB = xmysql.GetMySqlConn(emc.Alias)
		eventManager[emc.Alias].eventRetry = newEventRetry()
		eventManager[emc.Alias].startUp(emc.OpType, emc.KafkaConf)
	}
}

func (manager *EventManager) startUp(opType int, kafkaConf *mq.KafkaConf) {
	switch opType {
	case 1: // only event publish
		if err := manager.AutoMigrate(&pubEventObject{}); err != nil {
			panic(errors.WithStack(err))
		}
		mq.InitKafkaProducer(kafkaConf)

	case 2: // only event subscribe
		if err := manager.AutoMigrate(&subEventObject{}); err != nil {
			panic(errors.WithStack(err))
		}
		mq.InitKafkaConsumer(kafkaConf)

	case 3: // event publish and subscribe
		if err := manager.AutoMigrate(&pubEventObject{}); err != nil {
			panic(errors.WithStack(err))
		}
		if err := manager.AutoMigrate(&subEventObject{}); err != nil {
			panic(errors.WithStack(err))
		}

		mq.InitKafkaProducer(kafkaConf)
		mq.InitKafkaConsumer(kafkaConf)

	default:
		panic(errors.New("invalid table type"))
	}
}

func GetEventManager(alias string) *EventManager {
	return eventManager[alias]
}

func (manager *EventManager) SetRetryPolicy(delay time.Duration, retries int) {
	manager.eventRetry.setRetryPolicy(delay, retries)
}

func (manager *EventManager) push(event xevent.Event) {
	eventId := event.EventId()
	go func() {
		if err := manager.KafkaClent.SyncSend(event.Topic(), eventId, event.Content()); err != nil {
			return
		}

		if !manager.eventRetry.pubExist(event.Topic()) {
			return //未注册重试,结束
		}

		if err := manager.Where("identity = ?", eventId).Delete(&pubEventObject{}).Error; err != nil {
			//TODO
		}
	}()
}

func (manager *EventManager) Save(tx *gorm.DB, event *Event) (e error) {
	//for _, domainEvent := range entity.GetPubEvents() {
		if !manager.eventRetry.pubExist(event.Topic) {
			//continue //未注册重试,无需存储
		}

		model := pubEventObject{
			Identity: event.Id,
			Topic:    event.Topic,
			//Content:  string(domainEvent.Content()),
			Created:  time.Now(),
			Updated:  time.Now(),
		}
		if e = tx.Create(&model).Error; e != nil { //插入发布事件表。
			return
		}
	//}

	return
}

// InsertSubEvent .
func (manager *EventManager) InsertSubEvent(event xevent.Event) error {
	if !manager.eventRetry.subExist(event.Topic()) {
		return nil //未注册重试,无需存储
	}

	//content, err := event.Marshal()
	//if err != nil {
	//	return err
	//}

	model := subEventObject{
		Identity: event.EventId(),
		Topic:    event.Topic(),
		Content:  string(event.Content()),
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	return manager.Create(&model).Error //插入消费事件表。
}

// Retry .
func (manager *EventManager) Retry() {
	manager.eventRetry.retry()
}

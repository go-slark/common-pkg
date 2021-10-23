package domainevent

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Event struct {
	OpType int //1:insert 2: update 3:delete
	Topic string
	Id string
	Item map[string]interface{}
}

type EventTransaction struct {
	*gorm.DB //创建事务性db
	*EventManager
	*Event
}

func (et *EventTransaction) Execute(f func(db *gorm.DB) error) error {
	err := et.Transaction(func(tx *gorm.DB) error {
		//在这事务下执行业务数据保存和事件内容保存和回滚，事件发布成功后删除保存的事件内容，否则重试

		//保存业务数据
		e := f(tx)
		if e != nil {
			return errors.WithStack(e)
		}

		//保存事件内容
		e = et.EventManager.Save(tx, et.Topic, et.Item)
		if e != nil {
			return errors.WithStack(e)
		}

		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}

	//et.pushEvent()
	et.push(et.Event)
	return nil
}

//func (et *EventTransaction) pushEvent() {
//	for _, pubEvent := range et.GetPubEvents() {
//		et.push(pubEvent)
//	}
//}

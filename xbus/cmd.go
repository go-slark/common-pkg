package xbus

import (
	"context"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	"github.com/pkg/errors"
)

var commandBus *bus.CommandHandler

func InitCommandBus() {
	commandBus = bus.NewCommandHandler()
}

// RegisterCommandHandler handler实例需要实现 HandleCommand(context.Context, Command) error 这个接口
func RegisterCommandHandler(handler eh.CommandHandler, cmdType ...eh.CommandType) {
	for _, ct := range cmdType {
		err := commandBus.SetHandler(handler, ct)
		if err != nil {
			panic(errors.WithStack(err))
		}
	}
}

// RegisterCommand 注册命令
/*
    各命令需要实现三个函数
	func (c CreateInvite) AggregateID() uuid.UUID          { return c.ID }
	func (c CreateInvite) AggregateType() eh.AggregateType { return InvitationAggregateType }
	func (c CreateInvite) CommandType() eh.CommandType     { return CreateInviteCommand }
*/
func RegisterCommand(cmds ...eh.Command) {
	for _, cmd := range cmds {
		eh.RegisterCommand(func() eh.Command {return cmd})
	}
}

func getCommand(cmdType eh.CommandType) (eh.Command, error) {
	cmd, err := eh.CreateCommand(cmdType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return cmd, nil
}

//具体命令处理时实现HandleCommand这个接口，如写入数据库 返回响应等。。，成功之后 同时发布相关事件？
func handleCommand(ctx context.Context, cmd eh.Command) error {
	return commandBus.HandleCommand(ctx, cmd)
}

// HandleCommand 接收到路由进行需要进行命令处理,处理完之后就应该发布event, 然后消费端监听这个event
func HandleCommand(ctx context.Context, cmdType eh.CommandType) error {
	//根据url和cmd的映射关系找到相关的cmd，然后指向cmd实现了的HandleCommand接口
	//url和cmd的对应关系如h.Handle("/api/todos/check_all_items", httputils.CommandHandler(commandHandler, todo.CheckAllItemsCommand)) //CommandHandler 就是这里的HnadleCommand
	cmd, err := getCommand(cmdType)
	if err != nil {
		return errors.WithStack(err)
	}

	return handleCommand(ctx, cmd)
}

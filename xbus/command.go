package xbus

import (
	"context"
	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
	"github.com/pkg/errors"
	"github.com/smallfish-root/eventhorizon"
	"github.com/smallfish-root/eventhorizon/commandhandler/bus"
)

//command

func RegisterCommand(cmds ...eventhorizon.Command) {
	for _, cmd := range cmds {
		eventhorizon.RegisterCommand(func() eventhorizon.Command { return cmd })
	}
}

type CommandType = eventhorizon.CommandType
type Command struct {
	CmdType CommandType
}

func (c *Command) AggregateID() uuid.UUID {
	return uuid.Nil
}

func (c *Command) AggregateType() eventhorizon.AggregateType {
	return ""
}

func (c *Command) CommandType() eventhorizon.CommandType {
	return c.CmdType
}

func QueryCommand(cmdType CommandType) (eventhorizon.Command, error) {
	return eventhorizon.CreateCommand(cmdType)
}

//command handler

var cmdBus *bus.CommandHandler

func InitCmdBus() {
	cmdBus = bus.NewCommandHandler()
}

type CmdHandler = eh.CommandHandler
type CmdType = eh.CommandType
type CmdHandlerInfo struct {
	CmdType CmdType
	CmdHandler
}
//具体命令需要实现CmdHandler相关的函数
//
type Example struct {
	*CmdHandlerInfo
}

func (e *Example) HandleCommand(context.Context, Cmd) error {
	return nil
}

func New() *CmdHandlerInfo {
	e := &Example{}
	return &CmdHandlerInfo{
		CmdType:    "",
		CmdHandler: e,
	}
}
//

func RegisterCmdHandler(handlerInfos ...*CmdHandlerInfo) {
	for _, handlerInfo := range handlerInfos {
		err := cmdBus.SetHandler(handlerInfo.CmdHandler, handlerInfo.CmdType)
		if err != nil {
			panic(errors.WithStack(err))
		}
	}
}

type Cmd = eh.Command

func HandleCommand(ctx context.Context, cmd Cmd) error {
	return cmdBus.HandleCommand(ctx, cmd)
}

//变量的类型会涉及到使用looplab/enventhorizon的变量类型

//命令处理/事件发布基于event horizon(可能涉及到支持本地事务和发送失败定时重试-重新发布事件)，事件接收处理基于sarama处理。

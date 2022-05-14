package main

import (
	"fmt"
	"github.com/smallfish-root/common-pkg/xevent"
)

func main() {
	var event xevent.Event
	//订阅事件
	ch1 := make(chan int, 2) //订阅者1
	ch2 := make(chan int, 1) // 订阅者2
	event.Subscribe(ch2)
	sub, _ := event.Subscribe(ch1)
	sub.Unsubscribe()

	go func() {
		//发布事件
		event.Send(10000)
	}()

	fmt.Println(<-ch1, " ", <-ch2)
}

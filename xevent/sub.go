package xevent

type Subscription interface {
	Err() <-chan error
	Unsubscribe()
}

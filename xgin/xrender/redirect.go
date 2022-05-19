package xrender

import "github.com/gin-gonic/gin/render"

type Redirect struct {
	render.Redirect
	Code_ int
	Error error
}

func (r Redirect) Code() int {
	return r.Code_
}

func (r Redirect) Err() error {
	return r.Error
}

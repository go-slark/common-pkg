package xrender

import "github.com/gin-gonic/gin/render"

type Render interface {
	Code() int
	Err() error
	render.Render
}

type HttpCode struct {
	Code_ int
}

func (hc HttpCode) Code() int {
	return hc.Code_
}

type Error struct {
	error
}

func (e Error) Err() error {
	return e.error
}

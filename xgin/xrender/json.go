package xrender

import "github.com/gin-gonic/gin/render"

type JSON struct {
	HttpCode
	Trace
	Error
	render.JSON
}

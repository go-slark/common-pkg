package xgin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var requestId string

type config struct {
	builder   func() string
	requestId string
}

type Option func(*config)

func WithBuilder(b func() string) Option {
	return func(cfg *config) {
		cfg.builder = b
	}
}

func WithRequestId(requestId string) Option {
	return func(cfg *config) {
		cfg.requestId = requestId
	}
}

func BuildRequestId(opts ...Option) gin.HandlerFunc {
	cfg := &config{
		builder: func() string {
			return uuid.New().String()
		},
		requestId: "X-Request-ID",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx *gin.Context) {
		rid := ctx.GetHeader(cfg.requestId)
		if rid == "" {
			rid = cfg.builder()
		}
		requestId = cfg.requestId
		ctx.Header(cfg.requestId, rid)
		logrus.WithContext(context.WithValue(context.Background(), cfg.requestId, requestId))
	}
}

func GetRequestId(ctx *gin.Context) string {
	return ctx.Writer.Header().Get(requestId)
}

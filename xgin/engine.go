package xgin

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/smallfish-root/common-pkg/xerror"
	"github.com/smallfish-root/common-pkg/xlogger"
	"github.com/smallfish-root/common-pkg/xvalidator"
	"net/http"
)

type EngineParam struct {
	Env          string
	BaseUrl      string
	AccessLog    bool
	Pprof        bool
	ExcludePaths []string
	Routers      []func(r gin.IRouter)
	HandlerFunc  []gin.HandlerFunc
	ValidTrans   []xvalidator.ValidTrans
	xlogger.Logger
}

func SetEngine(param EngineParam) *gin.Engine {
	switch param.Env {
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
	xvalidator.RegisterCustomValidator(param.ValidTrans...)
	engine := gin.New()
	engine.Use(Recovery(func(ctx *gin.Context, err interface{}) {
		ctx.Render(http.StatusOK, Error(xerror.NewError(xerror.PanicCode, xerror.Panic, xerror.Panic).WithSurplus(err)))
	}))
	engine.Use(BuildRequestId())
	engine.Use(ErrLogger(param.Logger))
	if param.AccessLog {
		engine.Use(Logger(param.Logger, param.ExcludePaths...))
	}
	if param.Pprof {
		pprof.Register(engine)
	}
	engine.Use(param.HandlerFunc...)
	g := engine.Group(param.BaseUrl)
	for _, router := range param.Routers {
		router(g)
	}
	return engine
}

package xgin

import (
	"github.com/gin-gonic/gin"
	"github.com/smallfish-root/common-pkg/xvalidator"
)

func SetEngine(env string, vts ...xvalidator.ValidTrans) {
	switch env {
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
	xvalidator.SetValidatorToV9()
	xvalidator.RegisterCustomValidator(vts...)
}

package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	v1 "github.com/hpf0532/go-webhook/controller/v1"
	"github.com/hpf0532/go-webhook/extend/conf"
	"github.com/hpf0532/go-webhook/extend/logger"
	"time"
)

// InitRouter初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()
	// 注册zap相关中间件
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	gin.SetMode(conf.ServerConf.RunMode)
	// 跨域资源共享 CORS 配置
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  conf.CORSConf.AllowAllOrigins,
		AllowMethods:     conf.CORSConf.AllowMethods,
		AllowHeaders:     conf.CORSConf.AllowHeaders,
		ExposeHeaders:    conf.CORSConf.ExposeHeaders,
		AllowCredentials: conf.CORSConf.AllowCredentials,
		MaxAge:           conf.CORSConf.MaxAge * time.Hour,
	}))
	apiV1 := r.Group("api/v1")
	webHookController := new(v1.WebHookController)
	{
		apiV1.POST("/webhook", webHookController.HandleTask)
		apiV1.GET("/webhook", webHookController.HandleTask)
	}

	return r
}

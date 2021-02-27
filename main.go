package main

import (
	"fmt"
	"github.com/hpf0532/go-webhook/extend/conf"
	"github.com/hpf0532/go-webhook/extend/logger"
	"github.com/hpf0532/go-webhook/router"
	"log"
)

func main() {
	// 基本配置初始化
	conf.Setup()
	logger.Setup()
	r := router.InitRouter()
	log.Fatal(r.Run(fmt.Sprintf(":%d", conf.ServerConf.Port)))
}

package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hpf0532/go-webhook/extend/conf"
	"github.com/hpf0532/go-webhook/extend/logger"
	"github.com/hpf0532/go-webhook/utils"
	"io/ioutil"
)

// WebHookController webhook控制器
type WebHookController struct{}

func (wc *WebHookController) HandleTask(c *gin.Context) {
	//读取请求体
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.SugarLogger.Errorf("读取body错误: %s", err)
		return
	}
	logger.SugarLogger.Infof("当前请求体数据: %s", string(body))

	// 解析仓库名称
	repoName, err := utils.GetRepoName(string(body))
	if err != nil {
		logger.SugarLogger.Error(err)
		return
	}
	repoBranch, err := utils.GetRepoBranch(string(body))
	if err != nil {
		logger.SugarLogger.Error(err)
		return
	}
	webHookKey := fmt.Sprintf("%s/%s", repoName, repoBranch)
	logger.SugarLogger.Infof("webHookKey: %s", webHookKey)

	serverList, ok := conf.WebHookConf.WebHookMap[webHookKey]
	if !ok {
		logger.SugarLogger.Infof("没有匹配的仓库和分支, %s", webHookKey)
		return
	}
	fmt.Println(serverList)
	for _, host := range serverList {
		fmt.Println(host.Host)
	}

	c.JSON(200, gin.H{
		"message": "pong",
	})

}

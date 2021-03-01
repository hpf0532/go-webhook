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
		utils.ResponseFormat(c, 400, err.Error())
		return
	}
	repoBranch, err := utils.GetRepoBranch(string(body))
	if err != nil {
		logger.SugarLogger.Error(err)
		utils.ResponseFormat(c, 400, err.Error())
		return
	}
	webHookKey := fmt.Sprintf("%s/%s", repoName, repoBranch)
	logger.SugarLogger.Infof("webHookKey: %s", webHookKey)

	serverList, ok := conf.WebHookConf.WebHookMap[webHookKey]
	if !ok {
		logger.SugarLogger.Infof("没有匹配的仓库和分支, %s", webHookKey)
		utils.ResponseFormat(c, 200, "没有匹配的仓库和分支")
		return
	}
	if serverList != nil && len(serverList) > 0 {
		// 存在server，需要执行shell脚本
		for _, s := range serverList {
			if s.Script == "" {
				logger.SugarLogger.Warnf("脚本为空, webHookKey: %s", webHookKey)
				continue
			}
			logger.SugarLogger.Infof("开始执行脚本, %s", s.Script)
			fmt.Println(s)

		}
	}

	utils.ResponseFormat(c, 200, "success")

}

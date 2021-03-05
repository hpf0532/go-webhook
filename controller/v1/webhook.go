package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hpf0532/go-webhook/extend/command"
	"github.com/hpf0532/go-webhook/extend/conf"
	"github.com/hpf0532/go-webhook/extend/logger"
	"github.com/hpf0532/go-webhook/utils"
	"io/ioutil"
)

// WebHookController webhook控制器
type WebHookController struct{}

const EXECTIMEOUT = 3600

func (wc *WebHookController) HandleTask(c *gin.Context) {
	//读取请求体
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.SugarLogger.Errorf("读取body错误: %s", err)
		return
	}
	bodyStr := string(body)
	logger.SugarLogger.Infof("当前请求体数据: %s", bodyStr)

	// 解析仓库名称
	oName, repoName, err := utils.GetRepoName(bodyStr)
	if err != nil {
		logger.SugarLogger.Error(err)
		utils.ResponseFormat(c, 400, err.Error())
		return
	}
	repoBranch, err := utils.GetRepoBranch(bodyStr)
	if err != nil {
		logger.SugarLogger.Error(err)
		utils.ResponseFormat(c, 400, err.Error())
		return
	}
	webHookKey := fmt.Sprintf("%s/%s", repoName, repoBranch)
	logger.SugarLogger.Infof("webHookKey: %s", webHookKey)

	hook, ok := conf.WebHookConf.WebHookMap[webHookKey]
	if !ok {
		logger.SugarLogger.Infof("没有匹配的仓库和分支, %s", webHookKey)
		utils.ResponseFormat(c, 200, "没有匹配的仓库和分支")
		return
	}
	a := utils.GetLatestCommit(bodyStr)
	fmt.Println(a)
	if hook.Hook != nil && len(hook.Hook) > 0 {
		MsgInfo := utils.NewWebHookInfo()
		MsgInfo.Url = hook.Url
		MsgInfo.Comment = utils.GetLatestCommit(bodyStr)
		MsgInfo.Pusher = utils.GetPushUser(bodyStr)
		MsgInfo.RepoName = oName
		MsgInfo.Branch = repoBranch
		// 存在server，需要执行shell脚本
		go func() {
			command.Run(MsgInfo, hook, webHookKey, EXECTIMEOUT)
		}()
	}

	utils.ResponseFormat(c, 200, "success")

}

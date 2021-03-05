package message

import (
	"github.com/blinkbean/dingtalk"
	"github.com/hpf0532/go-webhook/extend/conf"
	"github.com/hpf0532/go-webhook/extend/logger"
	"regexp"
)

func DingTalkSend(title, msg string) {
	cli := dingtalk.InitDingTalkWithSecret(
		conf.DingTalkConf.AccessToken,
		conf.DingTalkConf.Secret,
	)
	markdown := []string{
		"## " + title,
		"---------",
		msg,
	}
	var err error
	at := conf.DingTalkConf.At
	if at == "all" {
		err = cli.SendMarkDownMessageBySlice(title, markdown, dingtalk.WithAtAll())
	} else {
		reg := regexp.MustCompile(`(\d{11})`)
		mobiles := reg.FindAllString(at, 2)
		if len(mobiles) > 0 {
			err = cli.SendMarkDownMessageBySlice(title, markdown, dingtalk.WithAtMobiles(mobiles))
		} else {
			err = cli.SendMarkDownMessageBySlice(title, markdown)
		}
	}

	if err != nil {
		logger.SugarLogger.Errorf("钉钉消息发送失败, %s", err)
	}

	logger.SugarLogger.Info("钉钉机器人推送成功")

}

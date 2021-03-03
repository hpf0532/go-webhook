package utils

import "github.com/hpf0532/go-webhook/extend/conf"

// 判断是否为远程主机
func IsRemote(host conf.HostConfig) (ok bool) {
	if host.Host == "" && host.Port == 0 && host.User == "" && host.Pwd == "" {
		return false
	} else {
		return true
	}
}

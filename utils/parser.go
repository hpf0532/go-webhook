package utils

import (
	"errors"
	"github.com/tidwall/gjson"
	"strings"
)

// 解析请求体中的仓库名称
func GetRepoName(data string) (string, error) {
	repo := gjson.Get(data, "repository.name")
	if repo.Exists() {
		return strings.ToLower(repo.String()), nil
	}
	repo = gjson.Get(data, "push_data.repository.name")
	if repo.Exists() {
		return strings.ToLower(repo.String()), nil
	}
	return "", errors.New("无法获取仓库名称")
}

// 解析请求体中的分支数据
func GetRepoBranch(data string) (string, error) {
	branch := gjson.Get(data, "ref")
	if !branch.Exists() {
		branch = gjson.Get(data, "push_data.ref")
		if !branch.Exists() {
			return "", errors.New("分支不存在")
		}
	}
	if strings.Contains(branch.String(), "/") {
		branchList := strings.Split(branch.String(), "/")
		return branchList[len(branchList)-1], nil
	}
	return branch.String(), nil
}

package utils

import (
	"errors"
	"fmt"
	"github.com/Albert-Zhan/httpc"
	"github.com/hpf0532/go-webhook/extend/conf"
	"github.com/tidwall/gjson"
	"net/http"
	"strings"
)

type WebHookInfo struct {
	Url      string
	Pusher   string
	RepoName string
	Comment  string
	Branch   string
}

func NewWebHookInfo() *WebHookInfo {
	return &WebHookInfo{}
}

// 解析请求体中的仓库名称
func GetRepoName(data string) (oname, name string, error error) {
	repo := gjson.Get(data, "repository.name")
	if repo.Exists() {
		return repo.String(), strings.ToLower(repo.String()), nil
	}
	repo = gjson.Get(data, "push_data.repository.name")
	if repo.Exists() {
		return repo.String(), strings.ToLower(repo.String()), nil
	}
	return "", "", errors.New("无法获取仓库名称")
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

// 解析push用户名
func GetPushUser(data string) (user string) {
	name := gjson.Get(data, "user_name") // gitlab
	if name.Exists() {
		return name.String()
	}
	name = gjson.Get(data, "pusher.name") // github
	if name.Exists() {
		return name.String()
	}
	name = gjson.Get(data, "pusher.username") // gogs
	if name.Exists() {
		return name.String()
	}
	name = gjson.Get(data, "push_data.user.name") // gitosc
	if name.Exists() {
		return name.String()
	}
	return ""
}

// 获取最新提交comment
func GetLatestCommit(data string) string {
	var comment string
	projectId := gjson.Get(data, "project_id")
	if !projectId.Exists() {
		return ""
	}
	branch, err := GetRepoBranch(data)
	if err != nil {
		return ""
	}
	fmt.Println(projectId, branch)
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits?ref_name=%s", conf.GitLabHost, projectId.String(), branch)
	fmt.Println(url)
	client := httpc.NewHttpClient()
	req := httpc.NewRequest(client)
	req.SetMethod("get").SetUrl(url)
	resp, body, err := req.SetHeader("PRIVATE-TOKEN", conf.GitLabToken).Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	commitList := gjson.Get(body, "#.message")
	commitList.ForEach(func(key, value gjson.Result) bool {
		c := value.String()
		if !strings.Contains(c, "Merge") {
			comment = c
			return false
		}
		return true
	})
	return strings.Replace(comment, "\n", "", -1)
}

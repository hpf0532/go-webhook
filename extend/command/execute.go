package command

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hpf0532/go-webhook/extend/conf"
	"github.com/hpf0532/go-webhook/extend/logger"
	"github.com/hpf0532/go-webhook/extend/message"
	"github.com/hpf0532/go-webhook/utils"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// result of the command execution
type ExecResult struct {
	Id             int
	Host           string
	Command        string
	LocalFilePath  string
	RemoteFilePath string
	Result         string
	StartTime      time.Time
	EndTime        time.Time
	Error          error
}

// ssh session
type HostSession struct {
	Username string
	Password string
	Hostname string
	Signers  []ssh.Signer
	Port     int
	Auths    []ssh.AuthMethod
}

// 生成ssh配置
func (exec *HostSession) GenerateConfig() ssh.ClientConfig {
	var auths []ssh.AuthMethod

	if len(exec.Password) != 0 {
		auths = append(auths, ssh.Password(exec.Password))
	} else {
		if len(exec.Auths) > 0 {
			auths = exec.Auths
		} else {
			auths = append(auths, ssh.PublicKeys(exec.Signers...))
		}

	}

	config := ssh.ClientConfig{
		User:            exec.Username,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// config.Ciphers = []string{"aes128-cbc", "3des-cbc"}

	return config
}

// 运行远端主机脚本，并返回结果
func (exec *HostSession) Exec(command string, config ssh.ClientConfig) *ExecResult {

	result := &ExecResult{
		Host:    exec.Hostname,
		Command: command,
	}

	client, err := ssh.Dial("tcp", exec.Hostname+":"+strconv.Itoa(exec.Port), &config)

	if err != nil {
		result.Error = err
		return result
	}

	session, err := client.NewSession()

	if err != nil {
		result.Error = err
		return result
	}

	defer session.Close()

	var b bytes.Buffer

	session.Stdout = &b
	var b1 bytes.Buffer
	session.Stderr = &b1
	start := time.Now()
	if err := session.Run(command); err != nil {
		result.Error = err
		result.Result = b1.String()
		return result
	}
	end := time.Now()
	result.Result = b.String()
	result.StartTime = start
	result.EndTime = end
	return result
}

// 加载秘钥
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// 秘钥配置
func GetAuthKeys(keys []string) []ssh.AuthMethod {
	methods := []ssh.AuthMethod{}
	for _, keyname := range keys {
		pkey := PublicKeyFile(keyname)
		if pkey != nil {
			methods = append(methods, pkey)
		}
	}
	return methods
}

// 密码配置
func GetAuthPassword(password string) []ssh.AuthMethod {
	return []ssh.AuthMethod{ssh.Password(password)}
}

func CommandBySSH(host conf.HostConfig, to int, wg *sync.WaitGroup) (*ExecResult, error) {
	defer wg.Done()
	var authKeys []ssh.AuthMethod
	timeout := time.After(time.Duration(to) * time.Second)
	execResultCh := make(chan *ExecResult, 1)
	// 密码不为空则使用密码连接
	if len(host.Pwd) > 0 {
		authKeys = GetAuthPassword(host.Pwd)

	} else {
		// 使用密钥连接
		keys := []string{
			os.Getenv("HOME") + "/.ssh/id_rsa",
			os.Getenv("HOME") + "/.ssh/id_dsa",
		}
		authKeys = GetAuthKeys(keys)
	}

	if len(authKeys) < 1 {
		logger.SugarLogger.Errorf("无法连接到%s, 没有匹配的key文件", host)
		return nil, errors.New("No such key.")
	}

	session := &HostSession{
		Hostname: host.Host,
		Username: host.User,
		Port:     host.Port,
		Password: host.Pwd,
		Auths:    authKeys,
	}
	go func() {
		sshResult := session.Exec(host.Script, session.GenerateConfig())
		execResultCh <- sshResult
	}()
	select {
	case res := <-execResultCh:
		sres := *res
		errorText := ""
		if sres.Error != nil {
			errorText += "Host " + sres.Host + " commond  exec error.\n" + "result info :" + sres.Result + "\nerror info :" + sres.Error.Error()
			logger.SugarLogger.Error(errorText)
		}
		if errorText != "" {
			return res, errors.New(errorText)
		} else {
			logger.SugarLogger.Infof("主机%s运行%s脚本完成, 运行结果为: \n%s", sres.Host, sres.Command, sres.Result)
			return res, nil
		}

	case <-timeout:
		logger.SugarLogger.Errorf("主机%s执行脚本%s超时", host.Host, host.Script)
		return &ExecResult{Command: host.Script, Error: errors.New("cmd time out")}, errors.New("cmd time out")
	}
}

func CommandLocal(cmd string, to int, wg *sync.WaitGroup) (ExecResult, error) {
	defer wg.Done()
	timeout := time.After(time.Duration(to) * time.Second)
	execResultCh := make(chan *ExecResult, 1)
	go func() {
		execResult := LocalExec(cmd)
		execResultCh <- &execResult
	}()
	select {
	case res := <-execResultCh:
		sres := *res
		errorText := ""
		if sres.Error != nil {
			errorText += "local commond  exec error.\n" + "result info :" + sres.Result + "\nerror info :" + sres.Error.Error()
			logger.SugarLogger.Error(errorText)
		}
		if errorText != "" {
			return sres, errors.New(errorText)
		} else {
			logger.SugarLogger.Infof("本地脚本%s运行完成, 运行结果为: \n%s", sres.Command, sres.Result)
			return sres, nil
		}

	case <-timeout:
		logger.SugarLogger.Errorf("本地脚本%s执行超时", cmd)
		return ExecResult{Command: cmd, Error: errors.New("cmd time out")}, errors.New("cmd time out")
	}

}

// 运行本地shell脚本
func LocalExec(cmd string) ExecResult {
	execResult := ExecResult{}
	execResult.StartTime = time.Now()
	execResult.Command = cmd
	execCommand := exec.Command("/bin/bash", "-c", cmd)
	var b bytes.Buffer
	execCommand.Stdout = &b
	var b1 bytes.Buffer
	execCommand.Stderr = &b1
	err := execCommand.Run()
	if err != nil {
		execResult.Error = err
		// execResult.ErrorInfo = err.Error()
		execResult.Result = b1.String()
		return execResult
	} else {
		execResult.EndTime = time.Now()
		execResult.Result = b.String()
		return execResult
	}
}

func Run(info *utils.WebHookInfo, hook *conf.HookConfig, key string, to int) {
	var err error
	var errList []string
	var wg sync.WaitGroup
	start := time.Now().Unix()
	for _, s := range hook.Hook {
		if s.Script == "" {
			logger.SugarLogger.Warnf("脚本为空, webHookKey: %s", key)
			continue
		}
		logger.SugarLogger.Infof("开始执行脚本, %s", s.Script)
		//go command.CommandLocal(s.Script, 3600)
		ok := utils.IsRemote(*s)
		wg.Add(1)
		if !ok {
			go func(s *conf.HostConfig) {
				_, err = CommandLocal(s.Script, to, &wg)
				if err != nil {
					errList = append(errList, err.Error())
				}
			}(s)
		} else {
			go func(s *conf.HostConfig) {
				_, err = CommandBySSH(*s, to, &wg)
				if err != nil {
					errList = append(errList, err.Error())
				}
			}(s)
		}

	}
	wg.Wait()
	end := time.Now().Unix()
	totalTime := end - start

	if len(errList) > 0 {
		errMsg := strings.Join(errList, "\n")
		if len(errMsg) > 300 {
			message.DingTalkSend(key, "任务发布失败，具体详情请联系管理员")
		} else {
			message.DingTalkSend(key, errMsg)
		}
		return
	}
	msg := fmt.Sprintf("- 提交人: %s  \n- 项目分支: %s  \n- Comment: %s  \n- 访问地址: %s  \n- 用时:%d秒  \n", info.Pusher, info.Branch, info.Comment, info.Url, totalTime)
	message.DingTalkSend(fmt.Sprintf("测试项目%s发布完成", info.RepoName), msg)

}

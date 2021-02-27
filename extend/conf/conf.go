package conf

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"time"
)

// server 服务基本配置结构
type server struct {
	RunMode      string        `mapstructure:"runMode"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

// Zap日志配置
type Zap struct {
	Level         string `mapstructure:"level"`
	Format        string `mapstructure:"format"`
	Prefix        string `mapstructure:"prefix"`
	Director      string `mapstructure:"director"`
	LinkName      string `mapstructure:"link-name"`
	ShowLine      bool   `mapstructure:"show-line"`
	EncodeLevel   string `mapstructure:"encode-level"`
	StacktraceKey string `mapstructure:"stacktrace-key"`
	LogInConsole  bool   `mapstructure:"log-in-console"`
}

// cors 跨域资源共享配置结构
type cors struct {
	AllowAllOrigins  bool          `mapstructure:"allowAllOrigins"`
	AllowMethods     []string      `mapstructure:"allowMethods"`
	AllowHeaders     []string      `mapstructure:"allowHeaders"`
	ExposeHeaders    []string      `mapstructure:"exposeHeaders"`
	AllowCredentials bool          `mapstructure:"allowCredentials"`
	MaxAge           time.Duration `mapstructure:"maxAge"`
}

// 主机脚本相关配置
type HostConfig struct {
	Host   string
	Port   int
	User   string
	Pwd    string
	Script string
}

// webhook配置
type SSHConfig struct {
	WebHookMap map[string][]*HostConfig
}

// ServerConf 服务基本配置
var ServerConf = &server{}

var ZapConf = &Zap{}

// CORSConf 跨域资源共享配置
var CORSConf = &cors{}

var WebHookConf = &SSHConfig{}

// 解析webhook相关配置参数
func ParseWebHookConf(conf map[string]interface{}) {
	WebHookConf.WebHookMap = make(map[string][]*HostConfig)
	for k, v := range conf {
		if item, ok := v.([]interface{}); ok {
			var hostList []*HostConfig
			for _, host := range item {
				var hostConf HostConfig = HostConfig{}
				if err := mapstructure.WeakDecode(host, &hostConf); err != nil {
					log.Fatalf("转换结构体失败, %s", err)
				}
				hostList = append(hostList, &hostConf)
			}
			WebHookConf.WebHookMap[k] = hostList

		} else {
			log.Fatal("配置文件格式错误")
		}
	}
	fmt.Println(WebHookConf.WebHookMap)
	for k, v := range WebHookConf.WebHookMap {
		fmt.Println(k)
		for i, j := range v {
			fmt.Println(i)
			fmt.Println("Host: ", j.Host)
			fmt.Println("Port: ", j.Port)
			fmt.Println("User: ", j.User)
			fmt.Println("Pwd: ", j.Pwd)
			fmt.Println("Script: ", j.Script)
		}
	}
}

// Setup 生成服务配置
func Setup() {

	viper.SetConfigType("yaml")
	// 读取配置文件内容
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	viper.UnmarshalKey("server", ServerConf)
	viper.UnmarshalKey("zap", ZapConf)
	viper.UnmarshalKey("cors", CORSConf)
	conf := viper.GetStringMap("webHookConfig")
	ParseWebHookConf(conf)
	viper.WatchConfig()
	// 动态加载webhook配置
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		conf := viper.GetStringMap("webHookConfig")
		ParseWebHookConf(conf)
	})
}

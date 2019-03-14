package config

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
)

// DefaultFile 默认配置文件
var DefaultFile = "~/.docker-debug/config.toml"

func init() {
	var err error
	var home string
	home, err = homedir.Dir()
	if err != nil {
		return
	}
	DefaultFile = fmt.Sprintf("%s/.docker-debug/config.toml", home)
}

// Config 配置
type Config struct {
	DefaultFile string
}

// DockerConfig docker 连接
type DockerConfig struct {
	Host string
	CA   string
}

type DebugConfig struct {
	DockerConfig    string
	DockerConfigMap map[string]DockerConfig
}


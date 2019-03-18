package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/zeromake/moby/opts"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
)

var configDir = ".docker-debug"

var configName = "config.toml"
var PathSeparator = string(os.PathSeparator)

// ConfigFile 默认配置文件
var ConfigFile = fmt.Sprintf(
	"~%s%s%s%s",
	PathSeparator,
	configDir,
	PathSeparator,
	configName,
)

func init() {
	var err error
	var home string
	home, err = homedir.Dir()
	if err != nil {
		return
	}
	configDir = fmt.Sprintf("%s%s%s", home, PathSeparator, configDir)
	ConfigFile = fmt.Sprintf("%s%s%s", configDir, PathSeparator, configName)
}

// DockerConfig docker 配置
type DockerConfig struct {
	Host         string
	TLS          bool
	CertDir      string
	CertPassword string
}

// Config 配置
type Config struct {
	Image               string
	//Command             []string
	Timeout             time.Duration
	DockerConfigDefault string
	DockerConfig        map[string]DockerConfig
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func LoadConfig() (*Config, error) {
	if !pathExists(ConfigFile) {
		return InitConfig()
	}
	config := &Config{}
	_, err := toml.DecodeFile(ConfigFile, config)
	return config, errors.WithStack(err)
}

func InitConfig() (*Config, error) {
	host, err := opts.ParseHost(false, false, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !pathExists(configDir) {
		err = os.Mkdir(configDir, 0755)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	config := &Config{
		Image:               "nicolaka/netshoot:latest",
		//Command:             []string{
		//	"sleep",
		//	"24h",
		//},
		Timeout:             time.Second * 10,
		DockerConfigDefault: "default",
		DockerConfig: map[string]DockerConfig{
			"default": DockerConfig{
				Host: host,
			},
		},
	}
	file, err := os.OpenFile(ConfigFile, os.O_CREATE | os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	encoder := toml.NewEncoder(file)
	defer func() {
		_ = file.Close()
	}()
	return config, encoder.Encode(config)
}

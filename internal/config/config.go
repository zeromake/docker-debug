package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/zeromake/docker-debug/pkg/opts"
	"os"
	"strings"
	"time"
	"github.com/mitchellh/go-homedir"
)

var configDir = ".docker-debug"

var configName = "config.toml"
var HOME = "~"
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
	var (
		home string
		err error
	)
	home, err = homedir.Dir()
	if err != nil {
		return
	}
	HOME = home
	configDir = fmt.Sprintf("%s%s%s", home, PathSeparator, configDir)
	ConfigFile = fmt.Sprintf("%s%s%s", configDir, PathSeparator, configName)
}

// DockerConfig docker 配置
type DockerConfig struct {
	Host         string `toml:"host"`
	TLS          bool `toml:"tls"`
	CertDir      string `toml:"cert_dir"`
	CertPassword string `toml:"cert_password"`
}

// Config 配置
type Config struct {
	Image string `toml:"image"`
	//Command             []string
	Timeout             time.Duration           `toml:"timeout"`
	DockerConfigDefault string                  `toml:"config_default"`
	DockerConfig        map[string]DockerConfig `toml:"config"`
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
	host, err := opts.ParseHost(false, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !pathExists(configDir) {
		err = os.Mkdir(configDir, 0755)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	dc := DockerConfig{
		Host: host,
	}
	if opts.IsWindows7 {
		paths := []string{
			HOME,
			".docker",
			"machine",
			"certs",
		}
		dc.TLS = true
		dc.CertDir = strings.Join(paths, PathSeparator)
	}
	config := &Config{
		Image: "nicolaka/netshoot:latest",
		//Command:             []string{
		//	"sleep",
		//	"24h",gaodingx_mysql_1
		//},frapsoft/htop
		Timeout:             time.Second * 10,
		DockerConfigDefault: "default",
		DockerConfig: map[string]DockerConfig{
			"default": dc,
		},
	}
	file, err := os.OpenFile(ConfigFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	encoder := toml.NewEncoder(file)
	defer func() {
		_ = file.Close()
	}()
	return config, encoder.Encode(config)
}

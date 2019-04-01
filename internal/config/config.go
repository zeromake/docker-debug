package config

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/zeromake/docker-debug/pkg/opts"
	"github.com/zeromake/docker-debug/version"
	"os"
	"strings"
	"time"
)

var configDir = ".docker-debug"

var configName = "config.toml"
// HOME system home path
var HOME = "~"
// PathSeparator path separator
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
		err  error
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
	TLS          bool   `toml:"tls"`
	CertDir      string `toml:"cert_dir"`
	CertPassword string `toml:"cert_password"`
}

func (c DockerConfig) String() string {
	s, _ := json.MarshalIndent(&c, "", "  ")
	return string(s)
}

// Config 配置
type Config struct {
	Version             string                  `toml:"version"`
	MountDir            string                  `toml:"mount_dir"`
	Image               string                  `toml:"image"`
	Timeout             time.Duration           `toml:"timeout"`
	DockerConfigDefault string                  `toml:"config_default"`
	DockerConfig        map[string]DockerConfig `toml:"config"`
}
// Save save to default file
func (c *Config) Save() error {
	file, err := os.OpenFile(ConfigFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	encoder := toml.NewEncoder(file)
	defer func() {
		_ = file.Close()
	}()
	return encoder.Encode(c)
}

// Load reload default file
func (c *Config) Load() error {
	_, err := toml.DecodeFile(ConfigFile, c)
	return errors.WithStack(err)
}

// PathExists path is has
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// LoadConfig load default file(not has init file)
func LoadConfig() (*Config, error) {
	if !PathExists(ConfigFile) {
		return InitConfig()
	}
	config := &Config{}
	_, err := toml.DecodeFile(ConfigFile, config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = MigrationConfig(config)
	return config, err
}

// InitConfig init create file
func InitConfig() (*Config, error) {
	host, err := opts.ParseHost(false, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !PathExists(configDir) {
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
		Version:             version.Version,
		Image:               "nicolaka/netshoot:latest",
		Timeout:             time.Second * 10,
		MountDir:            "/mnt/container",
		DockerConfigDefault: "default",
		DockerConfig: map[string]DockerConfig{
			"default": dc,
		},
	}
	file, err := os.OpenFile(ConfigFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	encoder := toml.NewEncoder(file)
	defer func() {
		_ = file.Close()
	}()
	return config, encoder.Encode(config)
}

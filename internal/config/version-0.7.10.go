package config

import (
	"github.com/blang/semver"
	"github.com/docker/docker/api"
)

// Up000710 update version 0.7.10
func Up000710(conf *Config) error {
	for _, c := range conf.DockerConfig {
		// 强制切换为 1.40
		if c.Version == "" || c.Version == api.DefaultVersion {
			c.Version = "1.40"
		}
	}
	return nil
}

func init() {
	v, err := semver.Parse("0.7.10")
	if err != nil {
		return
	}
	migrationArr = append(migrationArr, &migration{
		Up:      Up000710,
		Version: v,
	})
}

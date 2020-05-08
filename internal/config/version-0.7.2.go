package config

import (
	"github.com/blang/semver"
	"github.com/docker/docker/api"
)

// Up000702 update version 0.7.2
func Up000702(conf *Config) error {
	for _, c := range conf.DockerConfig {
		if c.Version == "" {
			c.Version = api.DefaultVersion
		}
	}
	return nil
}

func init() {
	v, err := semver.Parse("0.7.2")
	if err != nil {
		return
	}
	migrationArr = append(migrationArr, &migration{
		Up:      Up000702,
		Version: v,
	})
}

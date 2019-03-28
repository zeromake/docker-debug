package config

import (
	"github.com/blang/semver"
)

func Up000201(conf *Config) error {
	if conf.MountDir == "" {
		conf.MountDir = "/mnt/container"
	}
	return nil
}

func init() {
	v, err := semver.Parse("0.2.1")
	if err != nil {
		return
	}
	migrationArr = append(migrationArr, &migration{
		Up:      Up000201,
		Version: v,
	})
}

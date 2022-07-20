package config

import (
	"github.com/blang/semver"
	"time"
)

// Up000706 update version 0.7.6
func Up000706(conf *Config) error {
	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = time.Second * 3
	}
	return nil
}

func init() {
	v, err := semver.Parse("0.7.6")
	if err != nil {
		return
	}
	migrationArr = append(migrationArr, &migration{
		Up:      Up000706,
		Version: v,
	})
}

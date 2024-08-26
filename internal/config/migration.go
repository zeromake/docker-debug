package config

import (
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/zeromake/docker-debug/version"
	"sort"
	"strings"
)

type migration struct {
	Up      func(*Config) error
	Version semver.Version
}

var migrationArr []*migration

// MigrationConfig migration config version
func MigrationConfig(conf *Config) error {
	ver1 := version.Version
	var flag bool
	if strings.HasPrefix(ver1, "v") {
		ver1 = ver1[1:]
	}
	v1, err := semver.Parse(ver1)
	if err != nil {
		return nil
	}
	ver2 := conf.Version
	if strings.HasPrefix(ver2, "v") {
		ver2 = ver2[1:]
	}
	v2, err := semver.Parse(ver2)
	if err != nil {
		return errors.WithStack(err)
	}
	if strings.HasSuffix(conf.MountDir, "/") {
		flag = true
		l := len(conf.MountDir)
		conf.MountDir = conf.MountDir[:l-1]
	}
	if v2.LT(v1) {
		sort.Slice(migrationArr, func(i, j int) bool {
			return migrationArr[i].Version.LT(migrationArr[j].Version)
		})
		for _, m := range migrationArr {
			if v2.LT(m.Version) {
				err = m.Up(conf)
				if err != nil {
					return err
				}
			}
		}
		conf.Version = ver1
		return conf.Save()
	}
	if flag {
		return conf.Save()
	}
	return nil
}

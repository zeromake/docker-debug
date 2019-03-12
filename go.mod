module github.com/zeromake/docker-debug

go 1.12

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/Sirupsen/logrus v1.4.0
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-20170601211448-f5ec1e2936dc
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.3.3 // indirect
	github.com/google/go-cmp v0.2.1-0.20190228024137-c81281657ad9
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	golang.org/x/net v0.0.0-20190310074541-c10a0554eabf // indirect
	golang.org/x/sys v0.0.0-20190310054646-10058d7d4faa
	gotest.tools v0.0.0-20190311073145-20c9fe7f37cf
)

replace github.com/Sirupsen/logrus v1.4.0 => github.com/sirupsen/logrus v1.4.0

module github.com/TrilliumIT/docker-zfs-plugin

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/clinta/go-zfs v0.0.0-20181025145938-e5fe14d9dcb7
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-plugins-helpers v0.0.0-20200102110956-c9a8a2d92ccc
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.22.2
	golang.org/x/net v0.7.0 // indirect
)

replace github.com/docker/go-plugins-helpers => github.com/clinta/go-plugins-helpers v0.0.0-20200221140445-4667bb9f0ed5 // for shutdown

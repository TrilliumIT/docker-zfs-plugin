package main

import (
	"fmt"
	"os"

	zfsdriver "github.com/TrilliumIT/docker-zfs-plugin/zfs"
	"github.com/coreos/go-systemd/activation"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	version = "1.0.3"
)

func main() {

	app := cli.NewApp()
	app.Name = "docker-zfs-plugin"
	app.Usage = "Docker ZFS Plugin"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "dataset-name",
			Usage: "Name of the ZFS dataset to be used. It will be created if it doesn't exist.",
		},
	}
	app.Action = Run
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

// Run runs the driver
func Run(ctx *cli.Context) error {
	if ctx.String("dataset-name") == "" {
		return fmt.Errorf("zfs dataset name is a required field")
	}

	d, err := zfsdriver.NewZfsDriver(ctx.StringSlice("dataset-name")...)
	if err != nil {
		return err
	}
	h := volume.NewHandler(d)

	listeners, _ := activation.Listeners() // wtf coreos, this funciton never returns errors
	if len(listeners) == 0 {
		log.Debug("launching volume handler.")
		return h.ServeUnix("zfs", 0)
	}

	if len(listeners) > 1 {
		log.Warn("driver does not support multiple sockets")
	}

	l := listeners[0]
	log.WithField("listener", l.Addr().String()).Debug("launching volume handler")
	return h.Serve(l)
}

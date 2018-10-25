package main

import (
	"fmt"
	"os"

	"github.com/TrilliumIT/docker-zfs-plugin/zfs"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/urfave/cli"
)

const (
	version = "0.4.2"
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
		cli.BoolFlag{
			Name:  "enable-legacy-names",
			Usage: "Enable legacy (unqualified) names for the first specified dataset",
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
		return fmt.Errorf("ZFS Dataset name is a required field.")
	}

	d, err := zfsdriver.NewZfsDriver(ctx.StringSlice("dataset-name")...)
	if ctx.Bool("enable-legacy-names") {
		d.EnableLegacyNames()
	}
	if err != nil {
		return err
	}
	h := volume.NewHandler(d)
	h.ServeUnix("zfs", 0)

	return nil
}

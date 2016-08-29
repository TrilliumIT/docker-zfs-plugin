package main

import (
	"os"

	"github.com/TrilliumIT/docker-zfs-plugin/zfs"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/urfave/cli"
)

const (
	version = "0.2.2"
)

func main() {

	var flagDataset = cli.StringFlag{
		Name:  "dataset-name",
		Value: "",
		Usage: "Name of the ZFS dataset to be used. It will be created if it doesn't exist.",
	}

	app := cli.NewApp()
	app.Name = "docker-zfs-plugin"
	app.Usage = "Docker ZFS Plugin"
	app.Version = version
	app.Flags = []cli.Flag{
		flagDataset,
	}
	app.Action = Run
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

// Run runs the driver
func Run(ctx *cli.Context) {
	if ctx.String("dataset-name") == "" {
		panic("ZFS Dataset name is a required field.")
	}

	d, err := zfsdriver.NewZfsDriver(ctx.String("dataset-name"))
	if err != nil {
		panic(err)
	}
	h := volume.NewHandler(d)
	h.ServeUnix("root", "zfs")
}

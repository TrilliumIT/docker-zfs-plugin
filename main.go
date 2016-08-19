package main

import (
	"fmt"
	"os"

	"github.com/TrilliumIT/docker-zfs-plugin/zfs"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/urfave/cli"
)

const (
	version = "0.1"
)

func main() {

	var flagDataset = cli.StringFlag{
		Name:  "dataset-name",
		Value: "",
		Usage: "Name of the ZFS dataset to be used. It will be created if it doesn't exist.",
	}

	var flagMountpoint = cli.StringFlag{
		Name:  "mount-point",
		Value: "/var/lib/docker-volumes/zfs",
		Usage: "Mount point of your dataset. It will be created if it doesn't exist.",
	}

	app := cli.NewApp()
	app.Name = "docker-zfs-plugin"
	app.Usage = "Docker ZFS Plugin"
	app.Version = version
	app.Flags = []cli.Flag{
		flagDataset,
		flagMountpoint,
	}
	app.Action = Run
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func Run(ctx *cli.Context) {
	if ctx.String("dataset-name") == "" {
		panic("ZFS Dataset name is a required field.")
	}

	_, err := os.Stat(ctx.String("mount-point"))
	if err != nil {
		if os.IsNotExist(err) {
			err2 := os.MkdirAll(ctx.String("mount-point"), 0755)
			if err2 != nil {
				fmt.Errorf("Error creating mountpoint directory.")
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	d, err := zfsdriver.NewZfsDriver(ctx.String("dataset-name"), ctx.String("mount-point"))
	if err != nil {
		panic(err)
	}
	h := volume.NewHandler(d)
	h.ServeUnix("root", "zfs")
}

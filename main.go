package main

import (
	"fmt"
	"os"

	"github.com/TrilliumIT/docker-zfs-plugin/zfs"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/urfave/cli"
)

const (
	version = 0.1
)

func main() {

	var flagBaseDir = cli.StringFlag{
		Name:  "base-dir",
		Value: "/var/lib/docker-volumes",
		Usage: "Base directory where the driver root will be created.",
	}

	var flagRootDir = cli.StringFlag{
		Name:  "root-dir",
		Value: "zfs",
		Usage: "Relative name of the root directory for the driver. All volumes will be created in $BaseDir/$RootDir.",
	}

	app := cli.NewApp()
	app.Name = "docker-zfs-plugin"
	app.Usage = "Docker ZFS Plugin"
	app.Version = version
	app.Flags = []cli.Flag{
		flagBaseDir,
		flagRootDir,
	}
	app.Action = Run
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func Run(ctx *cli.Context) {
	_, err := FileInfo.Stat(ctx.String("base-dir"))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Error("Base directory must exist when launching docker-zfs-plugin.")
			return err
		}
	}
	d, err := vxlan.NewDriver(ctx.String("base-dir") + os.PathSeparator + ctx.String("root-dir"))
	if err != nil {
		panic(err)
	}
	h := volume.NewHandler(d)
	h.ServeUnix("root", "zfs")
}

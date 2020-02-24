package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	zfsdriver "github.com/TrilliumIT/docker-zfs-plugin/zfs"
	"github.com/coreos/go-systemd/activation"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	version         = "1.0.5"
	shutdownTimeout = 10 * time.Second
)

func main() {

	verbose := false
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
			Name:        "verbose",
			Usage:       "verbose output",
			Destination: &verbose,
		},
	}
	app.Action = Run
	app.Before = func(c *cli.Context) error {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}
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
	errCh := make(chan error)

	listeners, _ := activation.Listeners() // wtf coreos, this funciton never returns errors
	if len(listeners) > 1 {
		log.Warn("driver does not support multiple sockets")
	}
	if len(listeners) == 0 {
		log.Debug("launching volume handler.")
		go func() { errCh <- h.ServeUnix("zfs", 0) }()
	} else {
		l := listeners[0]
		log.WithField("listener", l.Addr().String()).Debug("launching volume handler")
		go func() { errCh <- h.Serve(l) }()
	}

	c := make(chan os.Signal)
	defer close(c)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	select {
	case err = <-errCh:
		log.WithError(err).Error("error running handler")
		close(errCh)
	case <-c:
	}

	toCtx, toCtxCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer toCtxCancel()
	if sErr := h.Shutdown(toCtx); sErr != nil {
		err = sErr
		log.WithError(err).Error("error shutting down handler")
	}

	if hErr := <-errCh; hErr != nil && !errors.Is(hErr, http.ErrServerClosed) {
		err = hErr
		log.WithError(err).Error("error in handler after shutdown")
	}

	return err
}

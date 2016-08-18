package zfs

import (
	"fmt"

	"github.com/docker/go-plugins-helpers/volume"
)

type Driver struct {
	volume.Driver
}

func NewDriver(dspath string) (Driver, error) {
	fmt.Println("test")

	return Driver{}, nil
}

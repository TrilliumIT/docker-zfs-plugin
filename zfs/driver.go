package zfsdriver

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/clinta/go-zfs"
	"github.com/docker/go-plugins-helpers/volume"
)

type ZfsDriver struct {
	volume.Driver
	rds *zfs.Dataset //root dataset
}

func NewZfsDriver(ds string, mp string) (*ZfsDriver, error) {
	log.SetLevel(log.DebugLevel)

	props := make(map[string]string)
	props["mountpoint"] = mp

	if !zfs.DatasetExists(ds) {
		rds, err := zfs.CreateDataset(ds, props)
		if err != nil {
			fmt.Errorf("Failed to create root dataset.")
			return nil, err
		}
		return &ZfsDriver{rds: rds}, nil
	}

	rds, err := zfs.GetDataset(ds)
	return &ZfsDriver{rds: rds}, err
}

func (zd *ZfsDriver) Create(req volume.Request) volume.Response {
	dsName := zd.rds.Name + "/" + req.Name

	if zfs.DatasetExists(dsName) {
		return volume.Response{Err: "Volume already exists."}
	}

	_, err := zfs.CreateDataset(dsName, make(map[string]string))
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (zd *ZfsDriver) List(req volume.Request) volume.Response {
	log.WithField("Requst", req).Debug("List()")
	var vols []*volume.Volume

	dsl, err := zd.rds.DatasetList()
	log.WithField("DatasetList", dsl).Debug("List()")
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	errStr := ""
	for _, ds := range dsl {
		mp, err := ds.GetMountpoint()
		if err != nil {
			errStr += "Failed to get mountpoint of dsl: " + ds.Name + " Error: " + err.Error() + "\n"
		}

		vols = append(vols, &volume.Volume{Name: volNameFromDsName(ds.Name), Mountpoint: mp})
	}

	return volume.Response{Volumes: vols, Err: errStr}
}

func (zd *ZfsDriver) Get(req volume.Request) volume.Response {
	dsName := zd.rds.Name + "/" + req.Name

	ds, err := zfs.GetDataset(dsName)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	mp, err := ds.GetMountpoint()
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Volume: &volume.Volume{Name: volNameFromDsName(ds.Name), Mountpoint: mp}, Err: ""}
}

func (zd *ZfsDriver) Remove(req volume.Request) volume.Response {
	dsName := zd.rds.Name + "/" + req.Name

	ds, err := zfs.GetDataset(dsName)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	err = ds.Destroy()
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (zd *ZfsDriver) Path(req volume.Request) volume.Response {
	ds := zd.Get(req)

	if ds.Err != "" {
		return ds
	}

	return volume.Response{Mountpoint: ds.Mountpoint, Err: ""}
}

func (zd *ZfsDriver) Mount(req volume.MountRequest) volume.Response {
	return zd.Path(volume.Request{Name: req.Name})
}

func (zd *ZfsDriver) Unmount(req volume.UnmountRequest) volume.Response {
	return volume.Response{Err: ""}
}

func (zd *ZfsDriver) Capabilities(req volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func volNameFromDsName(dsName string) string {
	volArr := strings.Split(dsName, "/")

	return volArr[len(volArr)-1]
}

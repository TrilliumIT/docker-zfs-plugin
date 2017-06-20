package zfsdriver

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/clinta/go-zfs"
	"github.com/docker/go-plugins-helpers/volume"
)

type ZfsDriver struct {
	volume.Driver
	rds *zfs.Dataset //root dataset
}

func NewZfsDriver(ds string) (*ZfsDriver, error) {
	log.SetLevel(log.DebugLevel)
	log.Debug("Creating new ZfsDriver.")

	if !zfs.DatasetExists(ds) {
		rds, err := zfs.CreateDataset(ds, make(map[string]string))
		if err != nil {
			log.Error("Failed to create root dataset.")
			return nil, err
		}
		return &ZfsDriver{rds: rds}, nil
	}

	rds, err := zfs.GetDataset(ds)
	return &ZfsDriver{rds: rds}, err
}

func (zd *ZfsDriver) Create(req volume.Request) volume.Response {
	log.WithField("Request", req).Debug("Create")

	dsName := zd.rds.Name + "/" + req.Name

	if zfs.DatasetExists(dsName) {
		return volume.Response{Err: "Volume already exists."}
	}

	_, err := zfs.CreateDataset(dsName, req.Options)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (zd *ZfsDriver) List(req volume.Request) volume.Response {
	log.WithField("Requst", req).Debug("List")
	var vols []*volume.Volume

	dsl, err := zd.rds.DatasetList()
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
	log.WithField("Request", req).Debug("Get")
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
	log.WithField("Request", req).Debug("Remove")
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
	log.WithField("Request", req).Debug("Path")
	res := zd.Get(req)

	if res.Err != "" {
		return res
	}

	return volume.Response{Mountpoint: res.Volume.Mountpoint, Err: ""}
}

func (zd *ZfsDriver) Mount(req volume.MountRequest) volume.Response {
	log.WithField("Request", req).Debug("Mount")

	return zd.Path(volume.Request{Name: req.Name})
}

func (zd *ZfsDriver) Unmount(req volume.UnmountRequest) volume.Response {
	log.WithField("Request", req).Debug("Unmount")
	return volume.Response{Err: ""}
}

func (zd *ZfsDriver) Capabilities(req volume.Request) volume.Response {
	log.WithField("Request", req).Debug("Capabilites")
	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func volNameFromDsName(dsName string) string {
	volArr := strings.Split(dsName, "/")

	return volArr[len(volArr)-1]
}

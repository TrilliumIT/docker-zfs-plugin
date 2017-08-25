package zfsdriver

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/clinta/go-zfs"
	"github.com/docker/go-plugins-helpers/volume"
	"strings"
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

func (zd *ZfsDriver) Create(req *volume.CreateRequest) error {
	log.WithField("Request", req).Debug("Create")

	dsName := zd.rds.Name + "/" + req.Name

	if zfs.DatasetExists(dsName) {
		return fmt.Errorf("Volume already exists.")
	}

	_, err := zfs.CreateDataset(dsName, req.Options)
	return err
}

func (zd *ZfsDriver) List() (*volume.ListResponse, error) {
	log.Debug("List")
	var vols []*volume.Volume

	dsl, err := zd.rds.DatasetList()
	if err != nil {
		return nil, err
	}

	errStr := ""
	for _, ds := range dsl {
		mp, err := ds.GetMountpoint()
		if err != nil {
			errStr += "Failed to get mountpoint of dsl: " + ds.Name + " Error: " + err.Error() + "\n"
		}

		vols = append(vols, &volume.Volume{Name: volNameFromDsName(ds.Name), Mountpoint: mp})
	}

	return &volume.ListResponse{Volumes: vols}, nil
}

func (zd *ZfsDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	mp, err := zd.getMP(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.GetResponse{Volume: &volume.Volume{Name: req.Name, Mountpoint: mp}}, nil
}

func (zd *ZfsDriver) getMP(name string) (string, error) {
	dsName := zd.rds.Name + "/" + name

	ds, err := zfs.GetDataset(dsName)
	if err != nil {
		return "", err
	}

	return ds.GetMountpoint()
}

func (zd *ZfsDriver) Remove(req *volume.RemoveRequest) error {
	log.WithField("Request", req).Debug("Remove")
	dsName := zd.rds.Name + "/" + req.Name

	ds, err := zfs.GetDataset(dsName)
	if err != nil {
		return err
	}

	return ds.Destroy()
}

func (zd *ZfsDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	log.WithField("Request", req).Debug("Path")
	mp, err := zd.getMP(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.PathResponse{Mountpoint: mp}, nil
}

func (zd *ZfsDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	log.WithField("Request", req).Debug("Mount")
	mp, err := zd.getMP(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.MountResponse{Mountpoint: mp}, nil
}

func (zd *ZfsDriver) Unmount(req *volume.UnmountRequest) error {
	log.WithField("Request", req).Debug("Unmount")
	return nil
}

func (zd *ZfsDriver) Capabilities() *volume.CapabilitiesResponse {
	log.Debug("Capabilites")
	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "local"}}
}

func volNameFromDsName(dsName string) string {
	volArr := strings.Split(dsName, "/")

	return volArr[len(volArr)-1]
}

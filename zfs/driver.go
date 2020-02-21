package zfsdriver

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/clinta/go-zfs"
	"github.com/docker/go-plugins-helpers/volume"
)

//ZfsDriver implements the plugin helpers volume.Driver interface for zfs
type ZfsDriver struct {
	volume.Driver
	rds []*zfs.Dataset //root dataset
}

//NewZfsDriver returns the plugin driver object
func NewZfsDriver(dss ...string) (*ZfsDriver, error) {
	log.SetLevel(log.DebugLevel)
	log.Debug("Creating new ZfsDriver.")

	zd := &ZfsDriver{}
	if len(dss) < 1 {
		return nil, fmt.Errorf("No datasets specified")
	}
	for _, ds := range dss {
		if !zfs.DatasetExists(ds) {
			_, err := zfs.CreateDatasetRecursive(ds, make(map[string]string))
			if err != nil {
				log.Error("Failed to create root dataset.")
				return nil, err
			}
		}
		rds, err := zfs.GetDataset(ds)
		if err != nil {
			log.Error("Failed to get root dataset.")
			return nil, err
		}
		zd.rds = append(zd.rds, rds)
	}

	return zd, nil
}

//Create creates a new zfs dataset for a volume
func (zd *ZfsDriver) Create(req *volume.CreateRequest) error {
	log.WithField("Request", req).Debug("Create")

	if zfs.DatasetExists(req.Name) {
		return fmt.Errorf("volume already exists")
	}

	_, err := zfs.CreateDatasetRecursive(req.Name, req.Options)
	return err
}

//List returns a list of zfs volumes on this host
func (zd *ZfsDriver) List() (*volume.ListResponse, error) {
	log.Debug("List")
	var vols []*volume.Volume

	for _, rds := range zd.rds {
		dsl, err := rds.DatasetList()
		if err != nil {
			return nil, err
		}
		for _, ds := range dsl {
			//TODO: rewrite this to utilize zd.getVolume() when
			//upstream go-zfs is rewritten to cache properties
			var mp string
			mp, err = ds.GetMountpoint()
			if err != nil {
				log.WithField("name", ds.Name).Error("Failed to get mountpoint from dataset")
				continue
			}
			vols = append(vols, &volume.Volume{Name: ds.Name, Mountpoint: mp})
		}
	}

	return &volume.ListResponse{Volumes: vols}, nil
}

//Get returns the volume.Volume{} object for the requested volume
//nolint: dupl
func (zd *ZfsDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	log.WithField("Request", req).Debug("Get")

	v, err := zd.getVolume(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.GetResponse{Volume: v}, nil
}

func (zd *ZfsDriver) getVolume(name string) (*volume.Volume, error) {
	ds, err := zfs.GetDataset(name)
	if err != nil {
		return nil, err
	}

	mp, err := ds.GetMountpoint()
	if err != nil {
		return nil, err
	}

	ts, err := ds.GetCreation()
	if err != nil {
		log.WithError(err).Error("Failed to get creation property from zfs dataset")
		return &volume.Volume{Name: name, Mountpoint: mp}, nil
	}

	return &volume.Volume{Name: name, Mountpoint: mp, CreatedAt: ts.Format(time.RFC3339)}, nil
}

func (zd *ZfsDriver) getMP(name string) (string, error) {
	ds, err := zfs.GetDataset(name)
	if err != nil {
		return "", err
	}

	return ds.GetMountpoint()
}

//Remove destroys a zfs dataset for a volume
func (zd *ZfsDriver) Remove(req *volume.RemoveRequest) error {
	log.WithField("Request", req).Debug("Remove")

	ds, err := zfs.GetDataset(req.Name)
	if err != nil {
		return err
	}

	return ds.Destroy()
}

//Path returns the mountpoint of a volume
//nolint: dupl
func (zd *ZfsDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	log.WithField("Request", req).Debug("Path")

	mp, err := zd.getMP(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.PathResponse{Mountpoint: mp}, nil
}

//Mount returns the mountpoint of the zfs volume
//nolint: dupl
func (zd *ZfsDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	log.WithField("Request", req).Debug("Mount")
	mp, err := zd.getMP(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.MountResponse{Mountpoint: mp}, nil
}

//Unmount does nothing because a zfs dataset need not be unmounted
func (zd *ZfsDriver) Unmount(req *volume.UnmountRequest) error {
	log.WithField("Request", req).Debug("Unmount")
	return nil
}

//Capabilities sets the scope to local as this is a local only driver
func (zd *ZfsDriver) Capabilities() *volume.CapabilitiesResponse {
	log.Debug("Capabilities")
	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "local"}}
}

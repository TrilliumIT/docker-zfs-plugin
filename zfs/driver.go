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
	rds         []*zfs.Dataset //root dataset
	legacyNames bool
}

func NewZfsDriver(dss ...string) (*ZfsDriver, error) {
	log.SetLevel(log.DebugLevel)
	log.Debug("Creating new ZfsDriver.")

	zd := &ZfsDriver{}
	if len(dss) < 1 {
		return nil, fmt.Errorf("No datasets specified")
	}
	for _, ds := range dss {
		if !zfs.DatasetExists(ds) {
			_, err := zfs.CreateDataset(ds, make(map[string]string))
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

func (zd *ZfsDriver) EnableLegacyNames() {
	zd.legacyNames = true
}

func (zd *ZfsDriver) dsName(name string) (string, error) {
	for _, ds := range zd.rds {
		if strings.HasPrefix(name, ds.Name+"/") && !strings.Contains(strings.TrimPrefix(name, ds.Name+"/"), "/") {
			return name, nil
		}
	}

	if zd.legacyNames && !strings.Contains(name, "/") {
		return zd.rds[0].Name + "/" + name, nil
	}

	return "", fmt.Errorf("Invalid dataset name: %v", name)
}

func (zd *ZfsDriver) Create(req *volume.CreateRequest) error {
	log.WithField("Request", req).Debug("Create")

	dsName, err := zd.dsName(req.Name)
	if err != nil {
		return err
	}

	if zfs.DatasetExists(dsName) {
		return fmt.Errorf("Volume already exists.")
	}

	_, err = zfs.CreateDataset(dsName, req.Options)
	return err
}

func (zd *ZfsDriver) List() (*volume.ListResponse, error) {
	log.Debug("List")
	var vols []*volume.Volume

	for i, rds := range zd.rds {
		dsl, err := rds.DatasetList()
		if err != nil {
			return nil, err
		}
		for _, ds := range dsl {
			mp, err := ds.GetMountpoint()
			if err != nil {
				return nil, err
			}
			vols = append(vols, &volume.Volume{Name: ds.Name, Mountpoint: mp})
			if i == 0 && zd.legacyNames {
				vols = append(vols, &volume.Volume{Name: strings.TrimPrefix(ds.Name, rds.Name), Mountpoint: mp})
			}
		}
	}

	return &volume.ListResponse{Volumes: vols}, nil
}

func (zd *ZfsDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	log.WithField("Request", req).Debug("Get")
	mp, err := zd.getMP(req.Name)
	if err != nil {
		return nil, err
	}

	return &volume.GetResponse{Volume: &volume.Volume{Name: req.Name, Mountpoint: mp}}, nil
}

func (zd *ZfsDriver) getMP(name string) (string, error) {
	dsName, err := zd.dsName(name)
	if err != nil {
		return "", err
	}

	ds, err := zfs.GetDataset(dsName)
	if err != nil {
		return "", err
	}

	return ds.GetMountpoint()
}

func (zd *ZfsDriver) Remove(req *volume.RemoveRequest) error {
	log.WithField("Request", req).Debug("Remove")
	dsName, err := zd.dsName(req.Name)
	if err != nil {
		return err
	}

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

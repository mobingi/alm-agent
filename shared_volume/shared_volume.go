package sharedvolume

import (
	"context"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	volumetypes "docker.io/go-docker/api/types/volume"
)

var (
	// DefarultMouhtPath is path to mount on container by default
	DefarultMouhtPath = "/mnt/storage"
)

// SharedVolume from serverconfig
type SharedVolume struct {
	Type       string `json:"type"`
	Identifier string `json:"id"`
	MountPath  string `json:"mountpath"`
}

// Interface is wrapper for docker volume
type Interface interface {
	Setup() error
	load()
}

// NullVolume is dummy volume for without volume
type NullVolume struct{}

// Setup do nothing
func (v *NullVolume) Setup() error {
	return nil
}

func (v *NullVolume) load() {
	return
}

// LocalVolume is simple volume
// inspect:
// [
//     {
//         "CreatedAt": "2018-10-23T07:36:23Z",
//         "Driver": "local",
//         "Labels": {},
//         "Mountpoint": "/var/lib/docker/volumes/test2/_data",
//         "Name": "test2",
//         "Options": {},
//         "Scope": "local"
//     }
// ]
type LocalVolume struct {
	Client *client.Client
	Name   string
	Path   string
	Volume *types.Volume
}

// Setup creates new or return exists volume
func (v *LocalVolume) Setup() error {
	v.load()
	vol, err := v.Client.VolumeCreate(
		context.Background(),
		volumetypes.VolumesCreateBody{Name: v.Name},
	)
	if err != nil {
		return err
	}

	v.Volume = &vol
	return nil
}

func (v *LocalVolume) load() {
	args := filters.NewArgs(
		filters.KeyValuePair{
			Key:   "name",
			Value: v.Name,
		},
	)
	vols, _ := v.Client.VolumeList(
		context.Background(),
		args,
	)
	if len(vols.Volumes) > 0 {
		v.Volume = vols.Volumes[0]
	}

	return
}

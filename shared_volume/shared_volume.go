package sharedvolume

import (
	"context"
	"fmt"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	volumetypes "docker.io/go-docker/api/types/volume"
)

// Interface is wrapper for docker volume
type Interface interface {
	Setup() (*types.Volume, error)
	status() *types.Volume
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
	client *client.Client
	name   string
}

// Setup creates new or return exists volume
func (v *LocalVolume) Setup() (*types.Volume, error) {
	vol, err := v.client.VolumeCreate(
		context.Background(),
		volumetypes.VolumesCreateBody{Name: v.name},
	)
	if err != nil {
		panic(err)
	}
	return &vol, nil
}

func (v *LocalVolume) status() *types.Volume {
	fmt.Printf("%#v\n", v.name)
	args := filters.NewArgs(
		filters.KeyValuePair{
			Key:   "name",
			Value: v.name,
		},
	)
	vols, _ := v.client.VolumeList(
		context.Background(),
		args,
	)
	if len(vols.Volumes) > 0 {
		fmt.Printf("%#v", vols.Volumes[0])
		return vols.Volumes[0]
	}

	return nil
}

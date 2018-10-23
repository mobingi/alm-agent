package sharedvolume

import (
	"context"
	"fmt"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	volumetypes "docker.io/go-docker/api/types/volume"
)

// SharedVolume from serverconfig
type SharedVolume struct {
}

// Interface is wrapper for docker volume
type Interface interface {
	Setup() error
	load()
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
	volume *types.Volume
}

// Setup creates new or return exists volume
func (v *LocalVolume) Setup() error {
	v.load()
	vol, err := v.client.VolumeCreate(
		context.Background(),
		volumetypes.VolumesCreateBody{Name: v.name},
	)
	if err != nil {
		return err
	}

	v.volume = &vol
	return nil
}

func (v *LocalVolume) load() {
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
		v.volume = vols.Volumes[0]
	}

	return
}

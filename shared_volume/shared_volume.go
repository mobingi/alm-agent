package sharedvolume

import (
	"context"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	volumetypes "docker.io/go-docker/api/types/volume"
)

type Interface interface {
	CreateVol() (*types.Volume, error)
}

type LocalVolume struct {
	client *client.Client
	name   string
}

func (v *LocalVolume) CreateVol() (*types.Volume, error) {
	vol, err := v.client.VolumeCreate(
		context.Background(),
		volumetypes.VolumesCreateBody{Name: v.name},
	)
	if err != nil {
		panic(err)
	}
	return &vol, nil
}

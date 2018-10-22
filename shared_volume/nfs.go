package sharedvolume

import (
	"context"

	client "docker.io/go-docker"
	volumetypes "docker.io/go-docker/api/types/volume"
)

// EFS example
// docker volume create --name test1  --opt type=nfs --opt device=":/" --opt o="xxx.xxx.xxx.xxx,nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport"

func createNFSVolume() error {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// volumetypes.VolumesCreateBody()

	cli.VolumeCreate(context.Background(), volumetypes.VolumesCreateBody{})

	return nil
}

// func (cli *Client) VolumeCreate(ctx context.Context, options volumetypes.VolumesCreateBody) (types.Volume, error)

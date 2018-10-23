package sharedvolume

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	volumetypes "docker.io/go-docker/api/types/volume"
)

var (
	mountopts = []string{
		"nfsvers=4.1",
		"rsize=1048576",
		"wsize=1048576",
		"hard",
		"timeo=600",
		"retrans=2",
		"noresvport",
	}
)

// https://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html
// file-system-id.efs.aws-region.amazonaws.com

// EFSVolume is Amazon EFS volume
// inspect example:
// [
//     {
//         "CreatedAt": "2018-10-23T07:35:45Z",
//         "Driver": "local",
//         "Labels": {},
//         "Mountpoint": "/var/lib/docker/volumes/efstest/_data",
//         "Name": "efstest",
//         "Options": {
//             "device": ":/",
//             "o": "addr=fs-xxxxxxxxx.efs.ap-northeast-1.amazonaws.com,nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport",
//             "type": "nfs"
//         },
//         "Scope": "local"
//     }
// ]
type EFSVolume struct {
	client *client.Client
	name   string
	efsid  string
}

// Setup creates new or return exists volume
func (v *EFSVolume) Setup() (*types.Volume, error) {
	o := fmt.Sprintf("%s,%s", v.efsAddr(), strings.Join(mountopts[:], ","))
	opts := map[string]string{
		"device": ":/",
		"type":   "nfs",
		"o":      o,
	}

	vol, err := v.client.VolumeCreate(
		context.Background(),
		volumetypes.VolumesCreateBody{
			Name:       v.name,
			Driver:     "local",
			DriverOpts: opts,
		},
	)
	if err != nil {
		panic(err)
	}
	return &vol, nil
}

func (v *EFSVolume) status() *types.Volume {
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

func (v *EFSVolume) getRegion() string {
	var METAENDPOINT = "http://169.254.169.254/"
	resp, err := http.Get(METAENDPOINT + "/latest/meta-data/placement/availability-zone")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%#v", err)
	}
	az := string(body)
	return az[0:(len(az) - 1)]
}

func (v *EFSVolume) efsAddr() string {
	// addr=fs-xxxxxxxx.efs.ap-northeast-1.amazonaws.com

	strAddr := fmt.Sprintf("addr=%s.efs.%s.amazonaws.com", v.efsid, v.getRegion())
	return strAddr
}

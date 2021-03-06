package sharedvolume

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/filters"
	volumetypes "docker.io/go-docker/api/types/volume"
	log "github.com/sirupsen/logrus"
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
	Client *client.Client
	Name   string
	EFSID  string
	Volume *types.Volume
}

// Setup creates new or return exists volume
func (v *EFSVolume) Setup() error {
	log.Infof("EFSVolume.Setup: %s", v.Name)
	err := v.load()
	if err != nil {
		return err
	}

	if v.Volume != nil {
		log.Debugf("found exists volume: %#v", v.Volume)
		return nil
	}

	log.Infof("EFSVolume.VolumeCreate: %s", v.Name)
	o := fmt.Sprintf("%s,%s", v.efsAddr(), strings.Join(mountopts[:], ","))
	opts := map[string]string{
		"device": ":/",
		"type":   "nfs",
		"o":      o,
	}

	vol, err := v.Client.VolumeCreate(
		context.Background(),
		volumetypes.VolumesCreateBody{
			Name:       v.Name,
			Driver:     "local",
			DriverOpts: opts,
		},
	)
	if err != nil {
		return err
	}

	v.Volume = &vol
	log.Debugf("created new volume: %#v", v.Volume)
	return nil
}

func (v *EFSVolume) load() error {
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
		err := v.verifyAndCleanup()
		if err != nil {
			log.Errorf("%#v", err)
			return err
		}
		return nil
	}

	return nil
}

// verify check parameters of exists volume.
// and clear if not correct
func (v *EFSVolume) verifyAndCleanup() error {
	log.Debug("Step: EFSVolume.verify")
	if v.Volume.Options != nil {
		if v.Volume.Options["type"] == "nfs" {
			if strings.HasPrefix(v.Volume.Options["o"], v.efsAddr()) {
				return nil
			}
		}
	}

	log.Warnf("found abnormal volume, tring to remove... : %#v", v.Volume)
	v.Volume = nil
	err := v.Client.VolumeRemove(
		context.Background(),
		v.Name,
		true,
	)

	if err != nil {
		return err
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

	strAddr := fmt.Sprintf("addr=%s.efs.%s.amazonaws.com", v.EFSID, v.getRegion())
	return strAddr
}

package bindata

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assets45fbe59f57b3389ef7dac8a7bdfc4ca770c3e99f = "[[container]]\n  name = \"alm-awslogs\"\n  image = \"mobingi/alm-awslogs\"\n  envfuncs = ['stack_id', 'instance_id']\n  volfuncs = ['logs_vol']\n"
var _Assets1b0bd80d2eaf5f4b5475dc8cbcd8406c2a2bdf7b = "[[container]]\n  name = \"mackerel\"\n  image = \"mackerel/mackerel-agent:latest\"\n  envfuncs = ['mackerel_envs']\n  volfuncs = ['mackerel_vol']\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{"/_data": []string{"sys_containers.toml", "addon_containers.toml"}, "/": []string{"_data"}}, map[string]*assets.File{
	"/": &assets.File{
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1510629064, 1510629064474443776),
		Data:     nil,
	}, "/_data": &assets.File{
		Path:     "/_data",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1509959672, 1509959672953932421),
		Data:     nil,
	}, "/_data/sys_containers.toml": &assets.File{
		Path:     "/_data/sys_containers.toml",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1504173724, 1504173724000000000),
		Data:     []byte(_Assets45fbe59f57b3389ef7dac8a7bdfc4ca770c3e99f),
	}, "/_data/addon_containers.toml": &assets.File{
		Path:     "/_data/addon_containers.toml",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1509959672, 1509959672954780137),
		Data:     []byte(_Assets1b0bd80d2eaf5f4b5475dc8cbcd8406c2a2bdf7b),
	}}, "")

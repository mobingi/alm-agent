package sharedvolume

// SharedVolume mounts on app container
type SharedVolume struct {
	Type      string `json:"type"`
	Identifer string `json:"id"`
	MountPath string `json:"mount_path"`
}

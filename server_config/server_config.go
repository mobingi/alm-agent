package serverConfig

type PubKey struct {
	PublicKey string
}

type Config struct {
	Image                string
	DockerHubUserName    string
	DockerHubPassword    string
	Code                 string
	CodeDir              string
	GitReference         string
	Ports                []int
	Users                map[string]*PubKey
	EnvironmentVariables map[string]string
}

package serverConfig

type Config struct {
	Image                string
	DockerHubUserName    string
	DockerHubPassword    string
	Code                 string
	CodeDir              string
	GitReference         string
	GitPrivateKey        string
	Ports                []int
	EnvironmentVariables map[string]string
}

package util

type alicloudProvider struct {
	*provider
}

func (p *alicloudProvider) MetadataEndpoint() string {
	return "http://100.100.100.200"
}

func (p *alicloudProvider) GetServerID() (string, error) {
	endpoint := p.MetadataEndpoint() + "/latest/meta-data/instance-id"

	sid, err := p.simpleHTTPGet(endpoint)
	if err != nil {
		return "", err
	}

	return sid, nil
}

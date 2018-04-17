package util

type awsProvider struct {
	*provider
}

func (p *awsProvider) MetadataEndpoint() string {
	return "http://169.254.169.254"
}

func (p *awsProvider) GetServerID() (string, error) {
	endpoint := p.MetadataEndpoint() + "/latest/meta-data/instance-id"

	sid, err := p.simpleHTTPGet(endpoint)
	if err != nil {
		return "", err
	}

	return sid, nil
}

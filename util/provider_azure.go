package util

type azureProvider struct {
	*provider
}

func (p *azureProvider) MetadataEndpoint() string {
	return "http://169.254.169.254/"
}

func (p *azureProvider) GetServerID() (string, error) {
	endpoint := p.MetadataEndpoint() + "metadata/instance/compute/vmId?api-version=2017-12-01&format=text"

	headers := map[string]string{
		"Metadata": "true",
	}

	sid, err := p.simpleHTTPGetWithHeader(endpoint, headers)
	if err != nil {
		return "", err
	}

	return sid, nil
}

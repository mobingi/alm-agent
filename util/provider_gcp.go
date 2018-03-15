package util

type gcpProvider struct {
	*provider
}

func (p *gcpProvider) MetadataEndpoint() string {
	return "http://metadata.google.internal/"
}

func (p *gcpProvider) GetServerID() (string, error) {
	endpoint := p.MetadataEndpoint() + "/computeMetadata/v1/instance/id"

	headers := map[string]string{
		"Metadata-Flavor": "Google",
	}

	sid, err := p.simpleHTTPGetWithHeader(endpoint, headers)
	if err != nil {
		return "", err
	}

	return sid, nil
}

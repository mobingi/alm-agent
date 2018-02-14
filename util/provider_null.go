package util

type nullProvider struct {
}

func (p *nullProvider) MetadataEndpoint() string {
	return ""
}

func (p *nullProvider) GetServerID() (string, error) {
	return "", nil
}

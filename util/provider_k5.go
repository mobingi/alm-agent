package util

import "encoding/json"

type k5Provider struct {
	*provider
}

func (p *k5Provider) MetadataEndpoint() string {
	return "http://169.254.169.254/"
}

func (p *k5Provider) GetServerID() (string, error) {
	endpoint := p.MetadataEndpoint() + "/openstack/latest/meta_data.json"

	body, err := p.simpleHTTPGet(endpoint)
	if err != nil {
		return "", err
	}

	sid, err := p.getUUID([]byte(body))
	if err != nil {
		return "", err
	}

	return sid, nil
}

func (p *k5Provider) getUUID(b []byte) (string, error) {
	var k k5
	if err := json.Unmarshal(b, &k); err != nil {
		return "", err
	}
	return k.UUID, nil
}

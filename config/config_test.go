package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	assert := assert.New(t)
	json := `
{
    "UserId": "5447826c870e7",
    "StackId": "mo-5447826c870e7-6ZIfI6Lf0-tk",
    "LogicalStackId" : "arn:aws:cloudformation:ap-northeast-1:963826138034:stack/mo-5447826c870e7-6ZIfI6Lf0-tk/a239a910-6ff4-11e6-8387-500c44f24ce6",
    "AccessKey": "AKIAJOC37T**********",
    "SecretKey": "yoWSII0QOF**********",
    "APIHost": "https://apidev.mobingi.com",
    "AuthorizationToken": "OHBXMGvtlecNNTF0u4KGB5WcZ05GVu",
    "LogBucket": "mocloud-customers",
    "ServerRole": "web",
    "Flag": "web01"
}
`

	c, err := Load([]byte(json))

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(c.UserID, "5447826c870e7")
	assert.Equal(c.Flag, "web01")
}

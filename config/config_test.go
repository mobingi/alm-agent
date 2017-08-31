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
    "APIHost": "https://apidev.mobingi.com",
    "AuthorizationToken": "OHBXMGvtlecNNTF0u4KGB5WcZ05GVu",
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

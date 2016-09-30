package config

import "testing"

func TestLoad(t *testing.T) {
	json := `
{
    "ServerConfigLocation": "s3://mocloud-customer/5447826c870e7/mo-5447826c870e7-6ZIfI6Lf0-tk/ServerConfig",
    "UserId": "5447826c870e7",
    "StackId": "mo-5447826c870e7-6ZIfI6Lf0-tk",
    "LogicalStackId" : "arn:aws:cloudformation:ap-northeast-1:963826138034:stack/mo-5447826c870e7-6ZIfI6Lf0-tk/a239a910-6ff4-11e6-8387-500c44f24ce6",
    "AccessKey": "AKIAJOC37T**********",
    "SecretKey": "yoWSII0QOF**********",
    "APIHost": "https://apidev.mobingi.com",
    "DBName": "none",
    "DBUsername": "none",
    "DBPassword": "none",
    "DBAddress": "none",
    "DBPort": "none",
    "EIP": "",
    "Redis": "none",
    "RedisAddress" : "none",
    "RedisPort" : "none",
    "AuthorizationToken": "OHBXMGvtlecNNTF0u4KGB5WcZ05GVu",
    "LogBucket": "mocloud-customers"
}
`

	c, err := Load([]byte(json))

	if err != nil {
		t.Fatal(err)
	}

	if c.UserID != "5447826c870e7" {
		t.Fatal("UserID is not correct")
	}
}

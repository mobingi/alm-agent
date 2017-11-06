package container

import (
	"testing"

	"docker.io/go-docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestCompareFuncsMackerelIcludesAll(t *testing.T) {
	assert := assert.New(t)
	d := &Docker{
		Envs: []string{
			"val1",
			"val2",
		},
	}
	c := &container.Config{
		Env: []string{
			"val1",
			"val2",
			"val3",
		},
	}
	actual := compareFuncs.Mackerel(d, c)
	assert.Equal(false, actual)
}

func TestCompareFuncsMackerelHasDiff(t *testing.T) {
	assert := assert.New(t)
	d := &Docker{
		Envs: []string{
			"val4",
			"val2",
		},
	}
	c := &container.Config{
		Env: []string{
			"val1",
			"val2",
			"val3",
		},
	}
	actual := compareFuncs.Mackerel(d, c)
	assert.Equal(true, actual)
}

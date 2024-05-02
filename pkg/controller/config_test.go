package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFromYaml(t *testing.T) {
	_, err := ReadFromYaml("server", "server.yml")
	assert := assert.New(t)
	assert.NotNil(err, "no this model")
	_, err = ReadFromYaml("trafficGate", "server.yml")
	assert.Nil(err, "successful read")
	_, err = ReadFromYaml("pipeline", "demo.yml")
	assert.Nil(err, "successful read")
}

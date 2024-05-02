package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFromYaml(t *testing.T) {
	_, err := ReadFromYaml("server")
	assert := assert.New(t)
	assert.NotNil(err, "no this model")
	_, err = ReadFromYaml("trafficGate")
	assert.Nil(err, "successful read")
	_, err = ReadFromYaml("pipeline")
	assert.Nil(err, "successful read")
}

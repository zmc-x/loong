package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockSpec struct {
	BaseSpec
	Mock string `json:"mock"`
}

var mockKind = &Kind{
	Name: "Mock",
	DefaultSpec: func() Spec {
		return &MockSpec{}
	},
	CreateInstance: func(_ Spec) Filter {
		return nil
	},
}

func TestNewSpec(t *testing.T) {
	assert := assert.New(t)
	_, err := NewSpec("pipeline-demo", "...")
	assert.NotNil(err, "json unmarshel error")

	jsonCfg := `
	{
		"name": "mock-demo",
		"kind": "Mock",
		"mock": "hello world"
	}
	`
	_, err = NewSpec("pipeline-demo", jsonCfg)
	assert.NotNil(err, "kinds not initialize")

	Registy(mockKind)
	defer ResetKinds(mockKind.Name)
	_, err = NewSpec("pipeline-demo", jsonCfg)
	assert.Nil(err, "success")
}
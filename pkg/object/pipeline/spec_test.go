package pipeline

import (
	"encoding/json"
	_ "loong/pkg/register"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cfg = `
{
	"name": "pipeline-ping",
	"kind": "Pipeline",
	"flow": [
	  {
		"filter": "validator-demo",
		"jumpIf": {
		  "invalid": "END",
		  "failureCode": "END"
		}
	  },
	  {
		"filter": "proxy-demo"
	  }
	],
	"filters": [
	  {
		"name": "validator-demo",
		"kind": "Validator",
		"headers": {
		  "Content-Type": {
			"values": [
			  "application/json"
			]
		  }
		},
		"jwt": {
		  "algorithm": "HS256",
		  "secret": "6d7973656372657"
		}
	  },
	  {
		"name": "proxy-demo",
		"kind": "Proxy",
		"pool": [
		  {
			"url": "http://127.0.0.1:9096"
		  },
		  {
			"url": "http://127.0.0.1:9095"
		  }
		],
		"loadBalance": {
		  "policy": "random"
		}
	  }
	]
}
`

func TestInitPipeline(t *testing.T) {
	str := []byte(cfg)
	assert := assert.New(t)
	_, err := InitPipeline(str)
	assert.Nil(err, "no success")
	flows := []string{"pipeline-ping:proxy-demo"}
	for _, flow := range flows {
		assert.NotNil(filters[flow], "no success")
	}
}

func TestSplitConfig(t *testing.T) {
	assert := assert.New(t)
	spec := Spec{}
	json.Unmarshal([]byte(cfg), &spec)
	_, err := splitCfg("proxy-demo", &spec)
	assert.Nil(err, "successful")
}

package controller

import (
	"loong/pkg/object/trafficgate"
	"os"

	"gopkg.in/yaml.v3"
)

func ReadFromYaml() (trafficgate.Config, error) {
	cfg := trafficgate.Config{}
	f, err := os.Open("/home/hellozmc/code/goproject/src/loong/temp/trafficgate/server.yml")
	if err != nil {
		return cfg, err
	}
	// unmarshal the yaml configuration file
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return cfg, err 
	}
	return cfg, nil
}


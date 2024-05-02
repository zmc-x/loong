package controller

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DirPath  = "/home/hellozmc/code/goproject/src/loong/temp"
	pipeline = "pipeline"
	server   = "trafficgate"
)

var ErrNoModel = errors.New("no this model")

func ReadFromYaml(key, fileName string) (any, error) {
	var cfg any
	var path string
	switch key {
	case "pipeline":
		path = filepath.Join(DirPath, pipeline, fileName)
	case "trafficGate":
		path = filepath.Join(DirPath, server, fileName)
	default:
		return nil, ErrNoModel
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	// unmarshal the yaml configuration file
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	// convert to json format
	return json.Marshal(cfg)
}

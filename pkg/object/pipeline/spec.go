package pipeline

import (
	"encoding/json"
	"errors"
	filter "loong/pkg/filters"
	"loong/pkg/supervisor"
	"net/http"
)

var filters map[string]filter.Filter = make(map[string]filter.Filter)

var (
	ErrNoConfig = errors.New("no this configuration file")
)

type Spec struct {
	supervisor.Meta `json:",inline"`

	// Flow represents the process of processing
	Flow []filterNode `json:"flow"`
	// Per filter configuration
	Filters []map[string]any `json:"filters"`
}

type filterNode struct {
	Filter string `json:"filter"`
}

// Register the pipeline with trafficgate
func (s *Spec) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Execute the handle in sequence
	for _, flow := range s.Flow {
		filterInstance := filters[flow.Filter]
		filterInstance.Handle(w, r)
	}
}

// create the filter
func InitPipeline(cfg any) (*Spec, error) {
	spec := Spec{}
	if err := json.Unmarshal(cfg.([]byte), &spec); err != nil {
		return nil, err
	}
	for _, node := range spec.Flow {
		nodeCfg, err := splitCfg(node.Filter, &spec)
		if err != nil {
			return nil, err
		}
		filterSpec, err := filter.NewSpec(spec.Name, nodeCfg)
		if err != nil {
			return nil, err
		}
		filters[node.Filter] = filter.Create(filterSpec)
		filters[node.Filter].Init()
	}
	return &spec, nil
}

// find the configuration of filter
func splitCfg(name string, rawCfg *Spec) (string, error) {
	for _, v := range rawCfg.Filters {
		if v["name"] == name {
			str, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(str), nil
		}
	}
	return "", ErrNoConfig
}
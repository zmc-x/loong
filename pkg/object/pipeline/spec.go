package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	filter "loong/pkg/filters"
	"loong/pkg/global"
	"loong/pkg/supervisor"
	"net/http"
)

const PipelineEND = "END"

var (
	PipelineMap map[string]*Spec = make(map[string]*Spec)
	// filters represent mappings between filters and corresponding modules in a pipeline.
	// key: pipeline:filter value: Filter
	filters     map[string]filter.Filter = make(map[string]filter.Filter)
	ErrNoConfig                          = errors.New("no this filter configuration file")
)

type Spec struct {
	supervisor.Meta `json:",inline"`
	Disable bool `json:"disable" validate:"boolean"`

	// Flow represents the process of processing
	Flow []filterNode `json:"flow"`
	// Per filter configuration
	Filters []map[string]any `json:"filters"`
}

type filterNode struct {
	Filter string `json:"filter"`
	// This mechanism is to enable certain modules to jump
	// to the corresponding position when they return a certain result,
	// similar to the jmp jump instruction in assembly.
	JumpIf map[string]string `json:"jumpIf,omitempty"`
}

// Register the pipeline with trafficgate
func (s *Spec) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Execute the handle in sequence
	statusCode := http.StatusOK
	lenf := len(s.Flow)
	for i := 0; i < lenf; i++ {
		flow := s.Flow[i]
		filterInstance := filters[s.Name+":"+flow.Filter]
		var res string
		res, tStatusCode := filterInstance.Handle(w, r)
		if statusCode != -1 {
			statusCode = tStatusCode
		}
		t := i
		if flow.JumpIf != nil {
			if v, ok := flow.JumpIf[res]; ok {
				if v == PipelineEND {
					break
				} else {
					for j := i + 1; j < lenf; j++ {
						if v == s.Flow[j].Filter {
							i = j - 1
							break
						}
					}
				}
			}
		}
		if t == i && res == "ratelimited" {
			break
		}
	}
	// Not through proxy
	if statusCode != -1 {
		w.WriteHeader(statusCode)
	}
}

func (s *Spec) Close() {
	lenf := len(s.Flow)
	for i := 0; i < lenf; i++ {
		flow := s.Flow[i]
		filterInstance := filters[s.Name+":"+flow.Filter]
		filterInstance.Close()
	}
}

// create the filter
func InitPipeline(cfg any) (*Spec, error) {
	spec := Spec{}
	if err := json.Unmarshal(cfg.([]byte), &spec); err != nil {
		return nil, err
	}
	if err := global.GlobalValidator.Struct(spec); err != nil {
		return nil, err
	}
	if spec.Disable {
		return nil, nil
	}
	// Check whether there are configurations with duplicate names
	if PipelineMap[spec.Name] != nil {
		return nil, fmt.Errorf("the pipeline of name %s already existes", spec.Name)
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
		key := spec.Name + ":" + node.Filter
		filters[key] = filter.Create(filterSpec)
		err = filters[key].Init()
		if err != nil {
			return nil, err
		}
	}
	PipelineMap[spec.Name] = &spec
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

// this function can reset some variables
// prevents errors during reloading
func Reset() {
	for _, pipeline := range PipelineMap {
		pipeline.Close()
	}
	PipelineMap = make(map[string]*Spec)
	filters = make(map[string]filter.Filter)
}
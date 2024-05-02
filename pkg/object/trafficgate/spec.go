package trafficgate

import "loong/pkg/supervisor"

// HTTPServer path configuration
type Spec struct {
	Path        string `json:"path"`
	// pipeline name
	Backend     string `json:"backend"`
}


type Config struct {
	supervisor.Meta `json:",inline"`
	Port  uint64 `json:"port"`
	Paths []Spec `json:"paths"`
}

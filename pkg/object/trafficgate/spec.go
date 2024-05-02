package trafficgate

import "loong/pkg/supervisor"

// HTTPServer path configuration
type Paths struct {
	Path string `json:"path"`
	// pipeline name
	Backend string `json:"backend"`
}

type Spec struct {
	supervisor.Meta `json:",inline"`
	Port            uint16  `json:"port"`
	Paths           []Paths `json:"paths"`
}

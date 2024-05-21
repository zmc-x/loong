package trafficgate

import "loong/pkg/supervisor"


var (
	// surMap check whether there is a duplicate trafficGate
	surMap  map[string]bool = make(map[string]bool)
	Servers []*Server
)

// HTTPServer path configuration
type Paths struct {
	Path string `json:"path" validate:"uri,required"`
	// pipeline name
	Backend  string `json:"backend" validate:"required"`
	IPFilter `json:"ipFilter,omitempty"`
	// Methods means this path supported request ways
	Methods []string `json:"methods,omitempty"`
}

type Spec struct {
	supervisor.Meta `json:",inline"`
	Disable bool `json:"disable" validate:"boolean"`
	IPFilter        `json:"ipFilter,omitempty"`
	Port            uint16  `json:"port" validate:"min=0,max=65535,required"`
	Paths           []Paths `json:"paths" validate:"dive"`
}

type IPFilter struct {
	// ip address that is allowed to access
	AllowIPs []string `json:"allowIPs" validate:"unique,dive,ip"`
	// ip address that is foribidden to access
	BlockIPs []string `json:"blockIPs" validate:"unique,dive,ip"`
}

// this function can reset some variables
// prevents errors during reloading
func Reset() {
	surMap = make(map[string]bool)
	Servers = nil
}
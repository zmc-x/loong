package trafficgate

// HTTPServer path configuration
type Spec struct {
	Path        string `yaml:"path"`
	Pool        []Host `yaml:"pool"`
	LoadBalance `yaml:"loadBalance,omitempty"`
}

type Host struct {
	Url string `yaml:"url"`
	// the field can be used to implement
	// a load balancing algorithm weight round-robin
	Weight int64 `yaml:"weight,omitempty"`
}

type LoadBalance struct {
	Policy string `yaml:"policy"`
}

type Config struct {
	Name  string `yaml:"name"`
	Kind  string `yaml:"kind"`
	Port  uint64 `yaml:"port"`
	Paths []Spec `yaml:"paths"`
}

package filters

import (
	"encoding/json"
	"errors"
	"loong/pkg/supervisor"
	"net/http"
)

var ErrNotFilter = errors.New("no this filter")

// kinds mappings of name and filter configurations
var kinds map[string]*Kind = make(map[string]*Kind)

func GetKind(name string) *Kind { return kinds[name] }

// create the filter
func Create(spec Spec) Filter {
	kind := kinds[spec.Kind()]
	if kind == nil {
		return nil
	}
	return kind.CreateInstance(spec)
}

func Registy(kind *Kind) {
	kinds[kind.Name] = kind
}

func ResetKinds(name string) {
	delete(kinds, name)
}

// Filter is common interface of filter
type Filter interface {
	// init this filter
	Init() error
	// Handle is handle logic
	Handle(http.ResponseWriter, *http.Request) (string, int)
}

type Kind struct {
	Name           string
	CreateInstance func(Spec) Filter
	DefaultSpec    func() Spec
}

// Spec is common data structure interface of filter
type Spec interface {
	// filter's name
	Name() string
	// filter's kind
	Kind() string
	// which pipeline belong with
	Pipeline() string
	// configuration file
	JSONConfig() string
	baseSpec() *BaseSpec
}

func NewSpec(pipelineName, rawCfg string) (Spec, error) {
	meta := supervisor.Meta{}
	err := json.Unmarshal([]byte(rawCfg), &meta)
	if err != nil {
		return nil, err
	}

	// get the specific spec
	kind := GetKind(meta.Kind)
	if kind == nil {
		return nil, ErrNotFilter
	}
	spec := kind.DefaultSpec()
	err = json.Unmarshal([]byte(rawCfg), spec)
	if err != nil {
		return nil, err
	}

	bs := spec.baseSpec()
	bs.pipeline = pipelineName
	bs.Meta = meta
	bs.jsonConfig = rawCfg

	return spec, nil
}

type BaseSpec struct {
	supervisor.Meta `json:",inline"`
	pipeline        string
	jsonConfig      string
}

func (b *BaseSpec) Name() string        { return b.Meta.Name }
func (b *BaseSpec) Kind() string        { return b.Meta.Kind }
func (b *BaseSpec) Pipeline() string    { return b.pipeline }
func (b *BaseSpec) JSONConfig() string  { return b.jsonConfig }
func (b *BaseSpec) baseSpec() *BaseSpec { return b }

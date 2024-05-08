package supervisor

// loong Api Gateway object meta data
type Meta struct {
	// loong API Gateway object kind
	Kind string `json:"kind" validate:"required"`
	// object name
	Name string `json:"name" validate:"required"`
}

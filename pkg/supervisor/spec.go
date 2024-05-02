package supervisor

// loong Api Gateway object meta data
type Meta struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	// RFC3339
	CreateAt string `json:"create_at"`
}

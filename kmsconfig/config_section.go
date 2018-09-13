package kmsconfig

type (
	// ConfigSection comment pending
	ConfigSection struct {
		Name  string
		Nodes map[string]ConfigNode
	}
)

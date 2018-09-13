package kmsconfig

type (
	// ConfigSection the top level node in the config.
	ConfigSection struct {
		Name  string
		Nodes map[string]ConfigNode
	}
)

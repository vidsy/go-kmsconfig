package kmsconfig

type (
	// ConfigNode a node in the config, a child of a
	// ConfigSection.
	ConfigNode struct {
		Name           string
		Value          interface{}
		EncryptedValue string
		Secure         bool
	}
)

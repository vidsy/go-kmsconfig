package kmsconfig

type (
	// ConfigNode comment pending
	ConfigNode struct {
		Name           string
		Value          interface{}
		EncryptedValue string
		Secure         bool
	}
)

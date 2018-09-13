package kmsconfig

type (
	// ConfigInterrogator is an interface for mocking config.
	ConfigInterrogator interface {
		Boolean(node string, key string) (bool, error)
		Integer(node string, key string) (int, error)
		String(node string, key string) (string, error)
		EncryptedString(node string, key string) (string, error)
		Environment() string
	}
)

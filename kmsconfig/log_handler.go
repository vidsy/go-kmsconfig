package kmsconfig

type (
	// LogHandler function that is called when lib
	// needs to log an action.
	LogHandler func(string)
)

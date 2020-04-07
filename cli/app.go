package cli

import (
	"fmt"
	"strings"

	"github.com/vidsy/go-kmsconfig/v5/kmsconfig"
)

type (
	// App represents the app for parsing confnig
	// value.
	App struct {
		config      *kmsconfig.Config
		NodeChild   string
		NodeSection string
	}
)

// NewApp creates a new App struct.
func NewApp(config *kmsconfig.Config, nodeName string) (*App, error) {
	nodeParts := strings.Split(nodeName, ".")

	if len(nodeParts) != 2 {
		return nil, fmt.Errorf(
			"Expected node name to be in format 'top_level_node.child_level_node', got: '%s'",
			nodeName,
		)
	}

	return &App{
		config:      config,
		NodeChild:   nodeParts[1],
		NodeSection: nodeParts[0],
	}, nil
}

// Value returns the value relating to the node
// key.
func (a App) Value() (interface{}, error) {
	return a.config.RawValue(a.NodeSection, a.NodeChild)
}

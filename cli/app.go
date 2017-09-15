package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vidsy/go-kmsconfig/kmsconfig"
)

type (
	// App represents the app for parsing confnig
	// value.
	App struct {
		config      *kmsconfig.Config
		NodeSection string
		NodeChild   string
		NodeType    string
	}
)

// NewApp creates a new App struct.
func NewApp(config *kmsconfig.Config, nodeName string, nodeType string) (*App, error) {
	nodeParts := strings.Split(nodeName, ".")

	if len(nodeParts) != 2 {
		return nil, fmt.Errorf(
			"Expected node name to be in format 'top_level_node.child_level_node', got: '%s'",
			nodeName,
		)
	}

	return &App{
		config:      config,
		NodeSection: nodeParts[0],
		NodeChild:   nodeParts[1],
		NodeType:    nodeType,
	}, nil
}

// Value returns the value relating to the node
// key.
func (a App) Value() (interface{}, error) {
	var value interface{}
	var err error

	switch a.NodeType {
	case "string":
		value, err = a.config.String(a.NodeSection, a.NodeChild)
	case "int":
		value, err = a.config.Integer(a.NodeSection, a.NodeChild)
	case "bool":
		value, err = a.config.Boolean(a.NodeSection, a.NodeChild)
	default:
		return nil, errors.New(
			"Node type not recognised, must be: 'string', 'int' or 'bool'",
		)
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

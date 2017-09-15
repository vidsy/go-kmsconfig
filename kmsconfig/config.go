package kmsconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const overrideEnvStructure = "VIDSY_VAR_%s_%s"

type (
	// ConfigInterrogator is an interface for mocking config.
	ConfigInterrogator interface {
		Boolean(node string, key string) (bool, error)
		Integer(node string, key string) (int, error)
		String(node string, key string) (string, error)
		EncryptedString(node string, key string) (string, error)
		Environment() string
	}

	// Config comment pending
	Config struct {
		data       map[string]map[string]map[string]interface{}
		Sections   map[string]ConfigSection
		Env        string
		Path       string
		KMSWrapper KMSWrapper
	}

	// ConfigSection comment pending
	ConfigSection struct {
		Name  string
		Nodes map[string]ConfigNode
	}

	// ConfigNode comment pending
	ConfigNode struct {
		Name           string
		Value          interface{}
		EncryptedValue string
		Secure         bool
	}
)

// NewConfig comment pending
func NewConfig(path string) *Config {
	env := environment()

	return &Config{
		Env:        env,
		Path:       path,
		KMSWrapper: NewKMSWrapper(),
	}
}

// Boolean returns a boolean cast value from a config node and key.
func (c Config) Boolean(node string, key string) (bool, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return false, err
	}

	return configNode.(bool), nil
}

// Environment returns the value of the .Env field.
func (c Config) Environment() string {
	return c.Env
}

// Integer comment pending
func (c Config) Integer(node string, key string) (int, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return 0, err
	}

	value := configNode.(float64)
	return int(value), nil
}

// String comment pending
func (c Config) String(node string, key string) (string, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return "", err
	}

	return configNode.(string), nil
}

// EncryptedString comment pending
func (c Config) EncryptedString(node string, key string) (string, error) {
	configNode, err := c.retrieve(node, key, true)
	if err != nil {
		return "", err
	}

	return configNode.(string), nil
}

// Load comment pending
func (c *Config) Load() error {
	path := c.generatePath()
	log.Printf(fmt.Sprintf("Loading config from '%s'", path))

	config, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(config), &c.data)
	if err != nil {
		return err
	}

	err = c.parse()
	if err != nil {
		return err
	}

	return nil
}

func (c Config) decryptSecureValue(key string, value string) (string, error) {
	log.Printf("Encrypted config value found for '%s', decrypting", key)
	decryptedValue, err := c.KMSWrapper.Decrypt(value)

	if err != nil {
		log.Printf("Failed to decrypt '%s'", key)
		return "", err
	}

	return decryptedValue, nil
}

func (c Config) overrideEnv(sectionValue string, nodeValue string) (string, bool) {
	environmentVariable := fmt.Sprintf(overrideEnvStructure, sectionValue, nodeValue)
	exists := os.Getenv(environmentVariable)

	if exists != "" {
		log.Printf("Override variable '%s' found", environmentVariable)
		return exists, true
	}

	return "", false
}

func (c *Config) parse() error {
	c.Sections = make(map[string]ConfigSection)

	for sectionKey, sectionValue := range c.data {
		configNodes := make(map[string]ConfigNode)

		section := ConfigSection{
			sectionKey,
			configNodes,
		}

		for nodeKey, nodeValue := range sectionValue {
			secure, _ := nodeValue["secure"].(bool)
			value := nodeValue["value"]
			encryptedValue := ""

			_, isString := nodeValue["value"].(string)
			_, isBool := nodeValue["value"].(bool)

			if isString {
				overrideEnvValue, envVarExists := c.overrideEnv(sectionKey, nodeKey)

				if envVarExists {
					value = overrideEnvValue
					encryptedValue = overrideEnvValue
				} else {
					if secure {
						decryptedValue, err := c.decryptSecureValue(nodeKey, value.(string))
						if err != nil {
							return err
						}

						encryptedValue = value.(string)
						value = decryptedValue
					}
				}
			}

			if isBool {
				overrideEnvValue, envVarExists := c.overrideEnv(sectionKey, nodeKey)

				if envVarExists {
					boolValue, err := strconv.ParseBool(overrideEnvValue)
					if err == nil {
						value = boolValue
					}
				}
			}

			node := ConfigNode{
				nodeKey,
				value,
				encryptedValue,
				secure,
			}

			configNodes[nodeKey] = node
		}

		c.Sections[sectionKey] = section
	}

	return nil
}

func (c Config) retrieve(node string, key string, encryptedValue bool) (interface{}, error) {
	section, sectionExists := c.Sections[node]

	if sectionExists {
		node, nodeExists := section.Nodes[key]

		if nodeExists {
			if encryptedValue {
				return node.EncryptedValue, nil
			}
			return node.Value, nil
		}

		return nil, fmt.Errorf("'%s' key doesn't exists on node '%s'", key, node.Name)
	}

	return nil, fmt.Errorf("The config node '%s' doesn't exist", node)
}

func environment() string {
	environment := "development"

	if os.Getenv("AWS_ENV") != "" {
		environment = os.Getenv("AWS_ENV")
	}

	return environment
}

func (c Config) generatePath() string {
	return fmt.Sprintf("%s/%s.json", c.Path, c.Env)
}

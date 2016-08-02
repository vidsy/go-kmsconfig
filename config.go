package kmsconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const OverrideEnvStructure = "VIDSY_VAR_%s_%s"

type (
	// Config comment pending
	Config struct {
		data       map[string]map[string]map[string]interface{}
		Sections   map[string]ConfigSection
		Env        string
		Path       string
		KMSWrapper aws.KMSWrapper
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
	return &Config{
		Path:       "./config",
		KMSWrapper: aws.NewKMSWrapper(),
	}
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
	c.Env = c.environment()

	filePath := fmt.Sprintf("%s/%s.json", c.Path, c.Env)
	log.Printf(fmt.Sprintf("Loading config from '%s'", filePath))

	config, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(config), &c.data)
}

// Parse comment pending
func (c *Config) Parse() error {
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

			if isString {
				overrideEnvValue, envVarExists := c.overrideEnv(sectionKey, nodeKey)

				if envVarExists {
					value = overrideEnvValue
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

func (c Config) overrideEnv(sectionValue string, nodeValue string) (string, bool) {
	environmentVariable := fmt.Sprintf(OverrideEnvStructure, sectionValue, nodeValue)
	exists := os.Getenv(environmentVariable)

	if exists != "" {
		log.Printf("Override variable '%s' found", environmentVariable)
		return exists, true
	}

	return "", false
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

func (c Config) retrieve(node string, key string, encryptedValue bool) (interface{}, error) {
	section, sectionExists := c.Sections[node]

	if sectionExists {
		node, nodeExists := section.Nodes[key]

		if nodeExists {
			if encryptedValue {
				return node.EncryptedValue, nil
			}
			return node.Value, nil
		} else {
			return nil, fmt.Errorf("'value' key doesn't exist on node '%s' doesn't exist", key)
		}

		// return nil, fmt.Errorf("'%s' key doesn't exists", key) TODO: this return statement is unreachable
	}

	return nil, fmt.Errorf("The config node '%s' doesn't exist", node)
}

func (c Config) environment() string {
	environment := "development"

	if os.Getenv("AWS_ENV") != "" {
		environment = os.Getenv("AWS_ENV")
	}

	return environment
}

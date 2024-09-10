package kmsconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const (
	overrideEnvStructure       = "VIDSY_VAR_%s_%s"
	configNodeName             = "config"
	configDurationTypeNodeName = "config_duration_type"
	configOmitField            = "-"
)

type Config struct {
	data       map[string]map[string]map[string]interface{}
	logHandler LogHandler
	Env        string
	KMSWrapper KMSWrapper
	Path       string
	Sections   map[string]ConfigSection
}

func NewConfig(path string, logHandler LogHandler) *Config {
	env := environment()

	return &Config{
		Env:        env,
		KMSWrapper: NewKMSWrapper(),
		logHandler: logHandler,
		Path:       path,
	}
}

func (c Config) Boolean(node string, key string) (bool, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return false, err
	}

	return configNode.(bool), nil
}

func (c Config) Environment() string {
	return c.Env
}

func (c Config) Integer(node string, key string) (int, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return 0, err
	}

	value := configNode.(float64)
	return int(value), nil
}

func (c *Config) Load() error {

	config, err := os.ReadFile(c.generatePath())
	if errors.Is(err, os.ErrNotExist) {
		return c.parseEnvsWithoutEncryption()
	}

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(config), &c.data)
	if err != nil {
		return err
	}

	return c.parse()
}

func (c *Config) LoadAndPopulate(config interface{}) error {
	if os.Getenv("VIDSY_VAR_CONFIG_EXCLUSIVELY_FROM_ENVIRONMENT") == "true" {
		return loadEnvConfig(config)
	}

	err := c.Load()
	if err != nil {
		return err
	}
	return c.Populate(config)
}

func (c Config) Populate(config interface{}) error {
	configPointer := reflect.ValueOf(config)
	if configPointer.Kind() != reflect.Ptr {
		return errors.New("Struct must be passed by reference")
	}

	configValue := configPointer.Elem()
	if configValue.NumField() == 0 {
		return errors.New("Expected struct to have >= 1 field, got 0")
	}

	for i := 0; i < configValue.NumField(); i++ {
		nodeFieldValue := configValue.Field(i)
		nodeFieldType := configValue.Type().Field(i)
		if nodeFieldValue.Kind() == reflect.Map {
			continue
		}
		if nodeFieldValue.NumField() == 0 {
			return errors.Errorf(
				"Struct '%s' should have 1 or more fields representing the second level of nesting in the config file, found no fields",
				nodeFieldType.Name,
			)
		}

		for j := 0; j < nodeFieldValue.NumField(); j++ {
			sectionFieldType := nodeFieldValue.Type().Field(j)
			sectionFieldValue := nodeFieldValue.Field(j)
			nodeTag := nodeFieldType.Tag.Get(configNodeName)
			sectionTag := sectionFieldType.Tag.Get(configNodeName)
			if nodeTag == configOmitField || sectionTag == configOmitField {
				continue
			}

			nodeData, err := c.retrieve(nodeTag, sectionTag, false)
			if err != nil {
				return errors.Wrapf(
					err,
					"Unabled to find config value for %s.%s",
					nodeTag,
					sectionTag,
				)
			}

			switch sectionFieldValue.Kind() {
			case reflect.Int64:
				var intType int64
				convertedValue := reflect.ValueOf(nodeData).Convert(reflect.TypeOf(intType))

				switch sectionFieldValue.Type().Name() {
				case "Duration":
					var duration time.Duration
					durationValue := convertedValue.Int()

					configDurationTypeTag := sectionFieldType.Tag.Get(configDurationTypeNodeName)
					switch configDurationTypeTag {
					case "microseconds":
						duration = time.Microsecond * time.Duration(durationValue)
					case "milliseconds":
						duration = time.Millisecond * time.Duration(durationValue)
					case "seconds":
						duration = time.Second * time.Duration(durationValue)
					case "minutes":
						duration = time.Minute * time.Duration(durationValue)
					case "hours":
						duration = time.Hour * time.Duration(durationValue)
					case "days":
						duration = (time.Hour * 24) * time.Duration(durationValue)
					default:
						return errors.Errorf(
							"Expected field of type time.Duration to have a struct tag '%s'",
							configDurationTypeNodeName,
						)
					}

					sectionFieldValue.Set(reflect.ValueOf(duration))
				default:
					sectionFieldValue.Set(convertedValue)
				}
			case reflect.Slice:
				slice, err := c.StringSlice(nodeTag, sectionTag)
				if err != nil {
					return err
				}

				sectionFieldValue.Set(reflect.ValueOf(slice))
			default:
				nodeDataValue := reflect.ValueOf(nodeData)
				if sectionFieldValue.Kind() != nodeDataValue.Kind() {
					return errors.Errorf(
						"Expected data type in field '%s' to be the same as the type in the config node, got: %s != %s",
						sectionFieldType.Name,
						sectionFieldValue.Kind(),
						nodeDataValue.Kind(),
					)
				}

				sectionFieldValue.Set(reflect.ValueOf(nodeData))
			}
		}
	}

	return nil
}

func (c Config) String(node string, key string) (string, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return "", err
	}

	return configNode.(string), nil
}

func (c Config) StringSlice(node string, key string) ([]string, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return nil, err
	}

	var values []string
	switch reflect.TypeOf(configNode).Kind() {
	case reflect.Slice:
		configNodeReflectedValue := reflect.ValueOf(configNode)
		for i := 0; i < configNodeReflectedValue.Len(); i++ {
			item := configNodeReflectedValue.Index(i).Elem()
			if item.Kind() != reflect.String {
				return nil, fmt.Errorf(
					"Mixed types in slice, expected all strings but got: %s",
					item.Kind(),
				)
			}

			values = append(
				values,
				item.String(),
			)
		}
	default:
		return nil, fmt.Errorf(
			"Expected underlying type to be a Slice, got: %s",
			reflect.TypeOf(configNode).Kind(),
		)
	}

	return values, nil
}

func (c Config) EncryptedString(node string, key string) (string, error) {
	configNode, err := c.retrieve(node, key, true)
	if err != nil {
		return "", err
	}

	return configNode.(string), nil
}

func (c Config) RawValue(node string, key string) (interface{}, error) {
	configNode, err := c.retrieve(node, key, false)
	if err != nil {
		return nil, err
	}

	return configNode, nil
}

func (c Config) decryptSecureValue(key string, value string) (string, error) {
	c.logHandler(
		fmt.Sprintf("Encrypted config value found for '%s', decrypting", key),
	)

	decryptedValue, err := c.KMSWrapper.Decrypt(value)

	if err != nil {
		return "", err
	}

	return decryptedValue, nil
}

func (c Config) overrideEnv(sectionValue string, nodeValue string) (string, bool) {
	environmentVariable := fmt.Sprintf(overrideEnvStructure, sectionValue, nodeValue)
	exists := os.Getenv(environmentVariable)

	if exists != "" {
		c.logHandler(
			fmt.Sprintf("Override variable '%s' found", environmentVariable),
		)
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

			overrideEnvValue, envVarExists := c.overrideEnv(sectionKey, nodeKey)
			if envVarExists {
				switch value.(type) {
				case string:
					value = overrideEnvValue
				case bool:
					boolValue, err := strconv.ParseBool(overrideEnvValue)
					if err != nil {
						return fmt.Errorf("error parsing env var override boolean value: %s for node %s.%s", err.Error(), sectionKey, nodeKey)
					}
					value = boolValue
				default:
					var jsonValue interface{}
					err := json.Unmarshal([]byte(overrideEnvValue), &jsonValue)
					if err != nil {
						return fmt.Errorf("error parsing env var override JSON value: %s for node %s.%s", err.Error(), sectionKey, nodeKey)
					}
					value = jsonValue
				}
			}

			if secure {
				encryptedStringValue, isString := value.(string)
				if !isString {
					return fmt.Errorf("secure value must be a string for node %s.%s", sectionKey, nodeKey)
				}
				decryptedValue, err := c.decryptSecureValue(nodeKey, encryptedStringValue)
				if err != nil {
					return fmt.Errorf("error decrypting secure value for node %s.%s: %s", sectionKey, nodeKey, err.Error())
				}
				encryptedValue = encryptedStringValue
				value = decryptedValue
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

func (c *Config) parseEnvsWithoutEncryption() error {
	c.Sections = make(map[string]ConfigSection)

	err := godotenv.Load(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "VIDSY_VAR_") {
			continue
		}

		partsA := strings.SplitN(strings.ToLower(strings.TrimPrefix(env, "VIDSY_VAR_")), "_", 2)
		if len(partsA) != 2 {
			continue
		}

		partsB := strings.SplitN(partsA[1], "=", 2)
		if len(partsB) != 2 {
			continue
		}

		sectionName := partsA[0]
		nodeName := partsB[0]
		nodeValue := partsB[1]

		if sectionName == "" || nodeName == "" || nodeValue == "" {
			continue
		}

		section, ok := c.Sections[sectionName]
		if !ok {
			section = ConfigSection{
				Name:  sectionName,
				Nodes: make(map[string]ConfigNode),
			}
		}

		section.Nodes[nodeName] = ConfigNode{
			Name:   nodeName,
			Value:  nodeValue,
			Secure: false,
		}

		c.Sections[sectionName] = section
	}

	return nil
}

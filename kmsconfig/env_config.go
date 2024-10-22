package kmsconfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func loadEnvConfig(config interface{}, kmsWrapper KMSWrapper) error {
	ctype := reflect.ValueOf(config)
	if ctype.Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer")
	}
	if ctype.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct pointer")
	}
	ctype = ctype.Elem()
	configMap, err := buildConfigMap(ctype)
	if err != nil {
		return err
	}

	return populateConfigFromEnv(configMap, kmsWrapper)
}

// buildConfigMap iterates over the fields of the config struct and builds a map of the field names to their values.
// the config struct is assumed to have two levels of fields, the first level being the "namespace" holding related fields,
// the second level being the actual configuration values.
// The function builds the map iterating over the "namespaces" and values, building the map keys as the corresponding environment
// variables holding the values.
// The map is then compared to the actual environment variables and the values are set accordingly.
func buildConfigMap(config reflect.Value) (map[string]reflect.Value, error) {
	configMap := make(map[string]reflect.Value)
	configType := config.Type()

	for i := 0; i < config.NumField(); i++ {
		namespaceValue := config.Field(i)

		namespaceTag := configType.Field(i).Tag.Get(configNodeName)
		if namespaceTag == "" {
			return nil, fmt.Errorf("config field %s has no config struct tag", configType.Field(i).Name)
		}
		if namespaceTag == configOmitField {
			continue
		}

		if configType.Field(i).Type.Kind() != reflect.Struct {
			return nil, fmt.Errorf("config field %s is not a struct", configType.Field(i).Name)
		}

		for j := 0; j < namespaceValue.NumField(); j++ {
			configFieldValue := namespaceValue.Field(j)
			configFieldType := namespaceValue.Type()

			fieldTag := configFieldType.Field(j).Tag.Get(configNodeName)
			if fieldTag == "" {
				return nil, fmt.Errorf("config field %s.%s has no config struct tag", configType.Field(i).Name, configFieldType.Field(j).Name)
			}
			if fieldTag == configOmitField {
				continue
			}

			envVar := fmt.Sprintf("VIDSY_VAR_%s_%s", strings.ToUpper(namespaceTag), strings.ToUpper(fieldTag))
			if _, ok := configMap[envVar]; ok {
				return nil, fmt.Errorf("the field %s.%s resolves to the environment variable %s which is already used by another field",
					configType.Field(i).Name, configFieldType.Field(j).Name, envVar)
			}

			configMap[envVar] = configFieldValue
		}
	}

	return configMap, nil
}

func populateConfigFromEnv(configMap map[string]reflect.Value, kmsWrapper KMSWrapper) error {
	envVars := map[string]string{}
	for _, envVar := range os.Environ() {
		v := strings.SplitN(envVar, "=", 2)
		envVars[v[0]] = v[1]
	}

	encryptedVariables := envVars["VIDSY_VAR_SECURED_ENVIRONMENT_VARIABLES"]
	encryptedVariablesMap := map[string]struct{}{}
	if len(encryptedVariables) > 0 {
		// use the existing code to parse these because why not?
		encryptedVariablesList := []string{}
		if err := assignEnvVarValue(reflect.ValueOf(&encryptedVariablesList).Elem(), encryptedVariables, "VIDSY_VAR_SECURED_ENVIRONMENT_VARIABLES"); err != nil {
			return err
		}
		for _, encryptedVariable := range encryptedVariablesList {
			encryptedVariablesMap[encryptedVariable] = struct{}{}
		}
	}

	// we expect to find all the environment variables from the config map
	for envVarName, value := range configMap {
		envValue, ok := envVars[envVarName]
		if !ok {
			return fmt.Errorf("environment variable %s not found", envVarName)
		}

		if _, ok := encryptedVariablesMap[envVarName]; ok {
			decryptedValue, err := kmsWrapper.Decrypt(envValue)
			if err != nil {
				return fmt.Errorf("error decrypting environment variable %s: %w", envVarName, err)
			}
			envValue = decryptedValue
		}

		return assignEnvVarValue(value, envValue, envVarName)
	}

	return nil
}

func assignEnvVarValue(value reflect.Value, envValue string, envVarName string) error {
	switch value.Kind() {
	case reflect.String:
		value.SetString(envValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing environment variable %s: %w", envVarName, err)
		}
		overflows := value.OverflowInt(intValue)
		if overflows {
			return fmt.Errorf("environment variable %s overflows the int type", envVarName)
		}

		value.SetInt(intValue)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intValue, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing environment variable %s: %w", envVarName, err)
		}
		overflows := value.OverflowUint(intValue)
		if overflows {
			return fmt.Errorf("environment variable %s overflows the uint type", envVarName)
		}

		value.SetUint(intValue)

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(envValue)
		if err != nil {
			return fmt.Errorf("error parsing environment variable %s: %w", envVarName, err)
		}
		value.SetBool(boolValue)

	case reflect.Slice:
		envValue = strings.TrimSpace(envValue)
		envValue = strings.TrimPrefix(envValue, "[")
		envValue = strings.TrimSuffix(envValue, "]")
		envValue = strings.ReplaceAll(envValue, "\"", "")

		envSliceValues := strings.Split(envValue, ",")
		slice := reflect.MakeSlice(value.Type(), 0, len(envSliceValues))
		for _, envSliceValue := range envSliceValues {
			appendedValue := reflect.New(value.Type().Elem())
			if err := assignEnvVarValue(appendedValue.Elem(), envSliceValue, envVarName); err != nil {
				return err
			}
			slice = reflect.Append(slice, appendedValue.Elem())
		}
		value.Set(slice)
	}

	return nil
}

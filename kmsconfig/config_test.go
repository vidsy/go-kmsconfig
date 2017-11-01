package kmsconfig_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vidsy/go-kmsconfig/kmsconfig"
)

func TestConfig(t *testing.T) {
	configLocation := "./fixtures/config"

	t.Run("LoadsConfigForDefaultEnvironment", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation)
		err := config.Load()
		assert.NoError(t, err)

		stringValue, err := config.String("app", "test_string")
		assert.NoError(t, err)
		assert.Equal(t, "foo", stringValue)

		intValue, err := config.Integer("app", "test_int")
		assert.NoError(t, err)
		assert.Equal(t, 1, intValue)

		boolValue, err := config.Boolean("app", "test_bool")
		assert.NoError(t, err)
		assert.Equal(t, true, boolValue)
	})

	t.Run("StringMap", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation)
		err := config.Load()
		assert.NoError(t, err)

		t.Run("ReturnsCorrectStringMap", func(t *testing.T) {
			stringMap, err := config.StringMap("app", "test_string_map")
			assert.NoError(t, err)

			test := map[string]map[string]string{
				"foo": map[string]string{
					"bar": "baz",
				},
			}

			assert.Equal(t, test, stringMap)
		})

		t.Run("ReturnsErrorIfNotMap", func(t *testing.T) {
			_, err := config.StringSlice("app", "test_string")
			assert.Error(t, err)
		})

		t.Run("ReturnsErrorIfKeyNotString", func(t *testing.T) {
			_, err := config.StringSlice("app", "test_string_map_wrong_type")
			assert.Error(t, err)
		})
	})

	t.Run(".StringSlice", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation)
		err := config.Load()
		assert.NoError(t, err)

		t.Run("ReturnsCorrectSlice", func(t *testing.T) {
			stringSlice, err := config.StringSlice("app", "test_string_slice")
			assert.NoError(t, err)
			assert.Equal(t, []string{"foo", "bar"}, stringSlice)
		})

		t.Run("ReturnsErrorIfNotSlice", func(t *testing.T) {
			_, err := config.StringSlice("app", "test_string")
			assert.Error(t, err)
		})

		t.Run("ReturnsErrorIfMixedValues", func(t *testing.T) {
			_, err := config.StringSlice("app", "test_string_slice_mixed_values")
			assert.Error(t, err)
		})
	})

	t.Run("LoadsConfigFromEnvironmentOverride", func(t *testing.T) {
		err := os.Setenv("AWS_ENV", "test")
		assert.NoError(t, err)

		config := kmsconfig.NewConfig(configLocation)
		err = config.Load()
		assert.NoError(t, err)

		stringValue, err := config.String("app", "test_string")
		assert.NoError(t, err)
		assert.Equal(t, "bar", stringValue)
	})

	t.Run("ErrorOnMissingEnvironment", func(t *testing.T) {
		err := os.Setenv("AWS_ENV", "foo")
		assert.NoError(t, err)

		config := kmsconfig.NewConfig(configLocation)
		err = config.Load()
		os.Unsetenv("AWS_ENV")
		assert.Error(t, err)
	})

	t.Run("NodeErrors", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation)
		err := config.Load()
		assert.NoError(t, err)

		t.Run("MissingTopLevelNode", func(t *testing.T) {
			_, err := config.String("foo", "bar")
			assert.Error(t, err)
		})

		t.Run("MissingChildNode", func(t *testing.T) {
			_, err := config.String("app", "foo")
			assert.Error(t, err)
		})
	})

	t.Run("OverrrideValue", func(t *testing.T) {
		err := os.Setenv("VIDSY_VAR_app_test_string", "baz")
		assert.NoError(t, err)

		config := kmsconfig.NewConfig(configLocation)
		err = config.Load()
		os.Unsetenv("VIDSY_VAR_app_string_value")

		stringValue, err := config.String("app", "test_string")
		assert.NoError(t, err)
		assert.Equal(t, "baz", stringValue)
	})
}

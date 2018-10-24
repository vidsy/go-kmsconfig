package kmsconfig_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vidsy/go-kmsconfig/kmsconfig"
)

func TestConfig(t *testing.T) {
	configLocation := "./fixtures/config"
	logHandler := func(message string) {}

	t.Run("LoadsConfigForDefaultEnvironment", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation, logHandler)
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

	t.Run(".StringSlice", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation, logHandler)
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

		config := kmsconfig.NewConfig(configLocation, logHandler)
		err = config.Load()
		assert.NoError(t, err)

		stringValue, err := config.String("app", "test_string")
		assert.NoError(t, err)
		assert.Equal(t, "bar", stringValue)
	})

	t.Run("ErrorOnMissingEnvironment", func(t *testing.T) {
		err := os.Setenv("AWS_ENV", "foo")
		assert.NoError(t, err)

		config := kmsconfig.NewConfig(configLocation, logHandler)
		err = config.Load()
		os.Unsetenv("AWS_ENV")
		assert.Error(t, err)
	})

	t.Run("NodeErrors", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation, logHandler)
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

		config := kmsconfig.NewConfig(configLocation, logHandler)
		err = config.Load()
		os.Unsetenv("VIDSY_VAR_app_test_string")

		stringValue, err := config.String("app", "test_string")
		assert.NoError(t, err)
		assert.Equal(t, "baz", stringValue)
	})

	t.Run(".Populate()", func(t *testing.T) {
		config := kmsconfig.NewConfig(configLocation, logHandler)
		err := config.Load()
		assert.NoError(t, err)

		t.Run("PopulatesStructCorrectly", func(t *testing.T) {
			var configStruct struct {
				App struct {
					TestBool        bool          `config:"test_bool"`
					TestString      string        `config:"test_string"`
					TestStringSlice []string      `config:"test_string_slice"`
					TestInt         int64         `config:"test_int"`
					TestTime        time.Duration `config:"test_time"`
				} `config:"app"`
			}

			err = config.Populate(&configStruct)
			assert.NoError(t, err)

			assert.Equal(t, "foo", configStruct.App.TestString)
			assert.Equal(t, int64(1), configStruct.App.TestInt)
			assert.Len(t, configStruct.App.TestStringSlice, 2)
			assert.True(t, configStruct.App.TestBool)
			assert.Equal(t, configStruct.App.TestTime, time.Duration(2))
		})

		t.Run("PopulatesStructProperlyWithOmittedFields", func(t *testing.T) {
			var configStruct struct {
				App struct {
					TestBool        bool     `config:"test_bool"`
					TestString      string   `config:"test_string"`
					TestStringSlice []string `config:"test_string_slice"`
					TestOmit        int64    `config:"-"`
				} `config:"app"`
			}

			configStruct.App.TestOmit = 10

			err = config.Populate(&configStruct)
			assert.NoError(t, err)

			assert.Equal(t, "foo", configStruct.App.TestString)
			assert.Len(t, configStruct.App.TestStringSlice, 2)
			assert.True(t, configStruct.App.TestBool)
			assert.Equal(t, int64(10), configStruct.App.TestOmit)
		})

		t.Run("ReturnsErrorIfPassedByValue", func(t *testing.T) {
			var configStruct struct {
				App struct {
					TestBool bool `config:"test_bool"`
				} `config:"app"`
			}

			err = config.Populate(configStruct)
			assert.Error(t, err)
		})

		t.Run("ReturnsErrorWithEmptyStruct", func(t *testing.T) {
			var configStruct struct{}
			err = config.Populate(&configStruct)
			assert.Error(t, err)
		})

		t.Run("ReturnsErrorWithNestedStructFiledWithNoValues", func(t *testing.T) {
			var configStruct struct {
				Foo struct{}
			}
			err = config.Populate(&configStruct)
			assert.Error(t, err)
		})

		t.Run("ReturnsErrorWithMissingNode", func(t *testing.T) {
			var configStruct struct {
				App struct {
					MissingField string `config:"missing_field"`
				} `config:"app"`
			}
			err = config.Populate(&configStruct)
			assert.Error(t, err)
		})

		t.Run("ReturnsErrorIfStructFieldTypeDifferentToTypeInJSONFile", func(t *testing.T) {
			var configStruct struct {
				App struct {
					TestString int `config:"test_string"`
				} `config:"app"`
			}
			err = config.Populate(&configStruct)
			assert.Error(t, err)
		})
	})
}

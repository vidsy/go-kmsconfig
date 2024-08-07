package kmsconfig_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vidsy/go-kmsconfig/v5/kmsconfig"
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

	t.Run("NoConfigFile", func(t *testing.T) {
		err := os.Setenv("AWS_ENV", "foo")
		assert.NoError(t, err)

		err = os.Setenv("VIDSY_VAR_FOO", "foo")
		assert.NoError(t, err)

		err = os.Setenv("VIDSY_VAR_FOO_BAR", "bar")
		assert.NoError(t, err)

		err = os.Setenv("VIDSY_VAR_FOO_BAR_BAZ", "baz")
		assert.NoError(t, err)

		config := kmsconfig.NewConfig(configLocation, logHandler)
		err = config.Load()
		assert.NoError(t, err)

		stringValue, err := config.String("foo", "bar")
		assert.NoError(t, err)
		assert.Equal(t, "bar", stringValue)

		stringValue, err = config.String("foo", "bar_baz")
		assert.NoError(t, err)
		assert.Equal(t, "baz", stringValue)

		os.Unsetenv("AWS_ENV")
		os.Unsetenv("VIDSY_VAR_FOO")
		os.Unsetenv("VIDSY_VAR_FOO_BAR")
		os.Unsetenv("VIDSY_VAR_FOO_BAR_BAZ")
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
		assert.NoError(t, err)
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
					TestBool             bool          `config:"test_bool"`
					TestInt              int64         `config:"test_int"`
					TestString           string        `config:"test_string"`
					TestStringSlice      []string      `config:"test_string_slice"`
					TestTimeDays         time.Duration `config:"test_time_days"         config_duration_type:"days"`
					TestTimeHours        time.Duration `config:"test_time_hours"        config_duration_type:"hours"`
					TestTimeMicroseconds time.Duration `config:"test_time_microseconds" config_duration_type:"microseconds"`
					TestTimeMilliseconds time.Duration `config:"test_time_milliseconds" config_duration_type:"milliseconds"`
					TestTimeMinutes      time.Duration `config:"test_time_minutes"      config_duration_type:"minutes"`
					TestTimeSeconds      time.Duration `config:"test_time_seconds"      config_duration_type:"seconds"`
				} `config:"app"`
			}

			err = config.Populate(&configStruct)
			assert.NoError(t, err)

			assert.Equal(t, "foo", configStruct.App.TestString)
			assert.Equal(t, int64(1), configStruct.App.TestInt)
			assert.Len(t, configStruct.App.TestStringSlice, 2)
			assert.True(t, configStruct.App.TestBool)
			assert.Equal(t, time.Duration(20*time.Microsecond), configStruct.App.TestTimeMicroseconds)
			assert.Equal(t, time.Duration(30*time.Millisecond), configStruct.App.TestTimeMilliseconds)
			assert.Equal(t, time.Duration(2*time.Second), configStruct.App.TestTimeSeconds)
			assert.Equal(t, time.Duration(5*time.Minute), configStruct.App.TestTimeMinutes)
			assert.Equal(t, time.Duration(10*time.Hour), configStruct.App.TestTimeHours)
			assert.Equal(t, time.Duration((time.Hour*24)*11), configStruct.App.TestTimeDays)
		})

		t.Run("ReturnsErrorIfConfigDurationTypeTagIsInvalid", func(t *testing.T) {
			var configStruct struct {
				App struct {
					TestTimeDay time.Duration `config:"test_time_days" config_duration_type:"error"`
				} `config:"app"`
			}

			err = config.Populate(&configStruct)
			assert.Error(t, err)
		})

		t.Run("ReturnsErrorIfConfigDurationTypeTagIsMissing", func(t *testing.T) {
			var configStruct struct {
				App struct {
					TestTimeDay time.Duration `config:"test_time_days"`
				} `config:"app"`
			}

			err = config.Populate(&configStruct)
			assert.Error(t, err)
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

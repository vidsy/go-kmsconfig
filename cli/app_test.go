package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vidsy/go-kmsconfig/cli"
	"github.com/vidsy/go-kmsconfig/kmsconfig"
)

func TestApp(t *testing.T) {
	t.Run("NewApp()", func(t *testing.T) {
		t.Run("ValidParameters", func(t *testing.T) {
			_, err := cli.NewApp(
				generateKMSConfig("foo", "baz"),
				"foo.bar",
				"string",
			)
			assert.NoError(t, err)
		})

		t.Run("ErrorWithInvalidNodeName", func(t *testing.T) {
			_, err := cli.NewApp(
				generateKMSConfig("foo", "baz"),
				"foo_bar",
				"string",
			)

			assert.Error(t, err)
		})
	})

	t.Run(".Value()", func(t *testing.T) {
		t.Run("StringValue", func(t *testing.T) {
			app, err := cli.NewApp(
				generateKMSConfig("foo", "baz"),
				"foo.bar",
				"string",
			)
			assert.NoError(t, err)

			value, err := app.Value()
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)
		})

		t.Run("IntValue", func(t *testing.T) {
			app, err := cli.NewApp(
				generateKMSConfig("foo", float64(1)),
				"foo.bar",
				"int",
			)
			assert.NoError(t, err)

			value, err := app.Value()
			assert.NoError(t, err)
			assert.Equal(t, 1, value)
		})

		t.Run("BoolValue", func(t *testing.T) {
			app, err := cli.NewApp(
				generateKMSConfig("foo", true),
				"foo.bar",
				"bool",
			)
			assert.NoError(t, err)

			value, err := app.Value()
			assert.NoError(t, err)
			assert.Equal(t, true, value)
		})

		t.Run("ErrorWhenKeyDoesntExist", func(t *testing.T) {
			app, err := cli.NewApp(
				generateKMSConfig("foo", "baz"),
				"foo.baz",
				"string",
			)
			assert.NoError(t, err)

			value, err := app.Value()
			assert.Error(t, err)
			assert.Nil(t, value)
		})
	})
}

func generateKMSConfig(nodeName string, nodeValue interface{}) *kmsconfig.Config {
	return &kmsconfig.Config{
		Sections: map[string]kmsconfig.ConfigSection{
			"foo": kmsconfig.ConfigSection{
				Name: "foo",
				Nodes: map[string]kmsconfig.ConfigNode{
					"bar": kmsconfig.ConfigNode{
						Name:  nodeName,
						Value: nodeValue,
					},
				},
			},
		},
	}

}

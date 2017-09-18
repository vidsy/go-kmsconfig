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
			)
			assert.NoError(t, err)
		})

		t.Run("ErrorWithInvalidNodeName", func(t *testing.T) {
			_, err := cli.NewApp(
				generateKMSConfig("foo", "baz"),
				"foo_bar",
			)

			assert.Error(t, err)
		})
	})

	t.Run(".Value()", func(t *testing.T) {
		t.Run("ValidValue", func(t *testing.T) {
			app, err := cli.NewApp(
				generateKMSConfig("foo", "bar"),
				"foo.bar",
			)
			assert.NoError(t, err)

			value, err := app.Value()
			assert.NoError(t, err)
			assert.Equal(t, "bar", value)
		})

		t.Run("ErrorWhenKeyDoesntExist", func(t *testing.T) {
			app, err := cli.NewApp(
				generateKMSConfig("foo", "baz"),
				"foo.baz",
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

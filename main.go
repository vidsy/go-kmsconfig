package main

import (
	"errors"
	"fmt"
	"os"

	"flag"
	"github.com/vidsy/go-kmsconfig/cli"
	"github.com/vidsy/go-kmsconfig/kmsconfig"
)

var (
	configPath = flag.String("path", "./config", "The path to the config folder")
	configNode = flag.String("node", "", "The node key to load, in the format: 'top_level_node.child_level_node'")
)

func main() {
	flag.Parse()

	config := kmsconfig.NewConfig(*configPath)
	err := config.Load()
	if err != nil {
		fatal(err)
	}

	if configNode == nil {
		fatal(
			errors.New("Please set -node parameter"),
		)
	}

	app, err := cli.NewApp(config, *configNode)
	if err != nil {
		fatal(err)
	}

	value, err := app.Value()
	if err != nil {
		fatal(err)
	}

	fmt.Println(value)
	os.Exit(0)
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(-1)
}

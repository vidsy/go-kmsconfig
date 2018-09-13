package main

import (
	"errors"
	"fmt"
	"log"
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

	logHandler := func(message string) {
		log.Println(message)
	}

	config := kmsconfig.NewConfig(*configPath, logHandler)
	err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if configNode == nil {
		log.Fatal(
			errors.New("Please set -node parameter"),
		)
	}

	app, err := cli.NewApp(config, *configNode)
	if err != nil {
		log.Fatal(err)
	}

	value, err := app.Value()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value)
	os.Exit(0)
}

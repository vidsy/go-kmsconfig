<h1 align="center">go-kmsconfig</h1>

<p align="center">
  JSON config with KMS encryption support.
</p>

## Setup

`go-kmsconfig` expects the following config structure:

```
- config
-- staging.json
-- qa.json
-- live.json
```

An example of a config file looks like:

```json

{
  "app": {
    "endpoint_url": {
      "value": "http://0.0.0.0:4569", 
      "secure": false
    }
  }
}
```

### Encrypted Values

Values can be encrypted with KMS and stored base64 encoded in the config. The consuming
service needs to have `Decrypt` permissions on the KMS key used to encrypt the value.

If the `secure` node is set to true for a child node then `go-kmsconfg` will attempt
to decrypt the value on load.

## Usage

```
glide get github.com/vidsy/go-kmsconfig
```

### Environment

By default, `go-kmsconfig` looks for `development.json` in the config folder provided
to `.NewConfig`.

For other environments, the following environment variable can be set:

```
AWS_ENV=staging
```

and `go-kmsconfig` will attempt to load `path_to_config/staging.json`.

#### Simple Example

```go
package main

import (
  "log"

  "github.com/vidsy/go-kmsconfig/kmsconfig"
)

func main() {
  parsedConfig := kmsconfig.NewConfig("./path_to_config_folder")

  err := parsedConfig.Load()
	if err != nil {
		log.Fatal(err)
	}

  configValue, err = parsedConfig.String("app", "some_config_node")
	if err != nil {
		return nil, err
	}

  log.Println("Config value for 'development.json' is: %s", configValue)
}
```

### Advanced Example

```go
package main

import (
  "log"

  "github.com/vidsy/go-kmsconfig/kmsconfig"
)

type (
  Config struct {
    App `config:"app"`
  }

  App struct {
    Counter          int64         `config:"counter"`
    Flag             bool          `config:"flag"`
    SleepDuration    time.Duration `config:"sleep_duration_seconds"  config_duration_type:"seconds"`
	}
)

func main() {
  var config Config
  err := config.Populate(&config)
  if err != nil {
		log.Fatal(err)
  }

  log.Println("SleepDuration value for 'development.json' is: %s", config.App.SleepDuration)
}
```

### Binary

For each tag a binary is also built that allows usage of the go-kmsconfg library from
a simple binary.

#### Setup

The binary can be found at:

```
https://s3-eu-west-1.amazonaws.com/go-kmsconfig.live.vidsy.co/${VERSION}/go-kmsconfig
```

> Where version is the tag you want the binary for.

#### Usage

```bash
$: ./go-kmsconfig --help 
```

#### Example

```bash
$: ./go-kmsconfig -path /path/to/config -node app.some_config_value
$: 23
```

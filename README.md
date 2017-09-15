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

An exmaple of a config file looks like:

```json

{
  "app": { // Top level node name
    "endpoint_url": { // Child node name
      "value": "http://0.0.0.0:4569", // Child node value
      "secure": false // Is the value encrypted with KMS?
    }
  }
}
```

### Encrypted Values

Values can be encrypted with KMS and stored base64 encoded in the config. The consuming
service needs to have `Decrypt` permissions on the KMS key used to encrypt the value.

Is the `secure` node is set to true for a child node then `go-kmsconfg` will attempt
to decrypt the value on load.

## Usage

```
glide get github.com/vidsy/go-kmsconfig
```

### Environment

By default, `go-kmsconfig` looks for `development.json` in the config folder provided
to `.NewConfig`.

For other environments, the following environment variable can be set:

`AWS_ENV=staging`

and `go-kmsconfig` will attempt to load `path_to_config/staging.json`.

#### Example

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

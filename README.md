# multiconfig

Easy to use automatic config structs. Easily load a config for your software by just using struct tags

## Features

- Specify your settings names in struct tags!
- Load variables from environment
- Load variables from json files/directories
- Load variables through argument cli flags
- Rename settings/automatically deals with nested fields
- Add usage to your flags through additional struct tags

## Description

`multiconfig` is a easy to use and lean golang module for managing configuration through declarative struct tagging.
Much as you would define struct tags for use in encoding/json or yaml files you should be able to do the same for command line
flags and environment variables. This is the core concept of `multiconfig`.

Available struct tags you can use for defining your configuration:

- `json`    ex. `json:"key"` -- the value of this tag will be used as the name of the config parameter associated with this struct field
- `env`     ex. `env:"key"`  -- will override json field name when parsing configs
- `usage`   ex. `usage:"key is used for secrets"`  -- used in the auto-generated usage message for --help, -h in the `flags` based config loader
- `arg`		ex. `arg:"key"` -- overrides the name of flag argument when using the `flags` based config loader

`multoconfig` uses a concept of a config *loader* to process different sources of configuration data. There are several loaders available by default:

- `NewMulti()` *loader* -- loads config data from multiple other *loader*s which are listed in order of reverse precedence (last one takes highest precedence)
- `NewJsonFile()` *loader* -- loads config data from a json file
- `NewJsonDirs()` *loader* -- loads config data from json files found in a list of directories
- `NewEnv(prefix)` *loader* -- loads config data from environment variables using the `env` or `json` struct tags for the name (plus optional `prefix`)
- `NewFlags()` *loader* -- loads config data using the `flag` standard library to generate cli flags and help usage output

You can specify default values for all of your settings at the *instance* level of your struct. Simply set the fields of your struct to the default values
you want to use when creating it before passing it to the *loader*s.

## Examples

### Example using NewMulti() (recommended way to use the library)

NewMulti() is the ideal way to use the library. It allows you to load multiple types of config in defined precedence order.

	package someapp
	
	import (
	    "github.com/fprintf/multiconfig"
	    "fmt"
	)
	
	type MySettings struct {
		// Specify different setting names for json, env and flags also specify usage string
		ServerName string `json:"server_name" env:"SERVER_NAME" arg:"servername" usage:"specify the name of the webserver"`
		// Env/Flag loaders will use 'json' tag if they have no name specific one
		Port int `json:"port" usage:"port for webserver"`
	}
	
	func main() {
		loader := multiconfig.NewMulti(
			multiconfig.NewEnv(""),
			multiconfig.NewFlags(),
		)
	
		// You can set default values by preloading the config instance before passing to `loader.Load()`
		settings := &MySettings{
			ServerName: "example.com",
			Port: 5000,
		}
		err := loader.Load(settings)
		if err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
	
		fmt.Printf("Loaded config: %#v", settings)
	}

### Example using NewEnv(prefix) on its own

NewEnv() loads config data into the struct from the environment variables passed to the program.
This generally means individual env var for each field in the struct. Specifying a prefix makes it easy
to remove repeated parts from your struct tag definitions.

	# config env vars
	MYAPP_SERVER_NAME="paradise-island.to"
	MYAPP_port=5000

	loader := multiconfig.NewEnv("MYAPP_")

	settings := &MySettings{}
	err := loader.Load(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	fmt.Printf("Loaded config: %#v", settings)

### Example using NewJsonFile(filepath)

NewJsonFile() will load the configuration data from a json file.

	# config.json
	{
		"server_name": "webserver.com",
		"port": 5000
	}

	# main.go
	loader := multiconfig.NewJsonFile("config.json")

	settings := &MySettings{}
	err := loader.Load(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	fmt.Printf("Loaded config: %#v", settings)

### Example using NewJsonDirs(dirpaths)

NewJsonDirs() will recursively search the list of directories, in order. Each file ending in .json
will be loaded into the given config struct. The last file found will have the most precedence on conflicting setting names.

>NOTE: Settings with the same name in each file will be overwritten. Settings with different names will be preserved.

	loader := multiconfig.NewJsonDirs([]string{
		"/home/user/.config/myapp",
		"/etc/myapp/config"
	})

	settings := &MySettings{}
	err := loader.Load(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	fmt.Printf("Loaded config: %#v", settings)
Also see cmd/testapp for a basic example usage

# multiconfig

Easy to use automatic config structs. Easily load a config for your software by just using struct tags

## Features

- Specify your settings names in struct tags!
- Load variables from environment
- Load variables from json files/directories
- Load variables through argument cli flags
- Rename settings/automatically deals with nested fields
- Add usage to your flags through additional struct tags

## Examples

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
	
		settings := &MySettings{}
		err := loader.Load(settings)
		if err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
	
		fmt.Printf("Loaded config: %#v", settings)
	}

Also see cmd/testapp for a basic example usage

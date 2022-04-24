package main

import (
	"log"

	"github.com/fprintf/multiconfig"
)

type TestConfig struct {
	ServerName    string `json:"server_name" env:"TEST_SERVER_NAME"`
	Port          int    `json:"port"`
	Pin           int8   `json:"pin"`
	ListenAddr    string `json:"listen_addr" arg:"bind" usage:"the address to bind our listener to"`
	RootDir       string `json:"root_dir"`
	User          string `json:"user"`
	Pass          string `json:"pass"`
	EnableLogging bool   `json:"enable_logging" usage:"enable logging"`
	Website       struct {
		Title   string
		Counter int
	}
	// TODO test this once we get array support working
	Names   []string       `json:"names" usage:"specify a list of names separate by comma or repeated flag calls"`
	Friends map[string]int `json:"friends" usage:"a map of friends name with an integer representing how much I like them"`
}

func main() {
	cfg := &TestConfig{
		ListenAddr: ":8080",
		Port:       22,
		User:       "testuser",
		Pass:       "nopassword",
		Names:      []string{"john", "nick"},
	}

	// Load config files in order of precedence
	loader := multiconfig.NewMultiLoader(
		multiconfig.NewEnv(""),
		multiconfig.NewFlags(),
	)

	err := loader.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%#v", cfg)
}

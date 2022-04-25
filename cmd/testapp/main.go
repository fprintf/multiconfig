package main

import (
	"encoding/json"
	"log"

	"github.com/fprintf/multiconfig"
)

type Embedded struct {
	Address string
	Age     int
}

type TestConfig struct {
	Embedded
	ServerName    string `json:"server_name" env:"TEST_SERVER_NAME"`
	Port          int    `json:"port"`
	Pin           int8   `json:"pin"`
	UnPin         uint8  `json:"upin"`
	ListenAddr    string `json:"listen_addr" arg:"bind" usage:"the address to bind our listener to"`
	RootDir       string `json:"root_dir"`
	User          string `json:"user"`
	Pass          string `json:"pass"`
	EnableLogging bool   `json:"enable_logging" usage:"enable logging"`
	Website       struct {
		Title   string
		Counter int
	}
	Names   []string       `json:"names" usage:"specify a list of names separate by comma or repeated flag calls"`
	Friends map[string]int `json:"friends" usage:"a map of friends name with an integer representing how much I like them"`
	File    string         `json:"file" argtype:"positional"`
	File2   string         `json:"file2" argtype:"positional"`
	Args    []string       `json:"args" argtype:"positional"`
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
	loader := multiconfig.NewMulti(
		multiconfig.NewEnv(""),
		multiconfig.NewFlags(),
	)

	err := loader.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(cfg, "", "    ")
	log.Println(string(out))
}

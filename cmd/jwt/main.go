package main

import (
	"encoding/json"
	"log"

	"github.com/fprintf/multiconfig"
)

type JwtConfig struct {
	Args []string `json:"args" argtype:"positional"`
}

func main() {
	// Load config files in order of precedence
	loader := multiconfig.NewMulti(
		multiconfig.NewEnv(""),
		multiconfig.NewFlags(),
	)

	cfg := &JwtConfig{}
	err := loader.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(cfg, "", "    ")
	log.Println(string(out))
}

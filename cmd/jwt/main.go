package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fprintf/multiconfig"
	"github.com/golang-jwt/jwt"
)

type JwtConfig struct {
	Files []string `json:"files" argtype:"positional"`
}

// OpenFileLogError opens a file with the given name and returns
// an io.ReadCloser. If the file name is an empty string, e.g "", then
// os.Stdin will be returned
func OpenFileLogError(file string) io.ReadCloser {
	if file == "" {
		return os.Stdin
	}

	infile, err := os.Open(file)
	if err != nil {
		fmt.Printf("failed to open %s: %v\n", file, err)
		return nil
	}
	return infile
}

// CloseFileLogError closes the io.ReadCloser and prints an error if there was one
func CloseFileLogError(infile io.ReadCloser) {
	err := infile.Close()
	if err != nil {
		fmt.Printf("error %v\n", err)
	}
}

// Token is a wrapper around jwt.Token for debugging/printing purposes
type Token struct {
	*jwt.Token
}

func (t *Token) String() string {
	val, _ := json.MarshalIndent(t, "", "    ")
	return string(val)
}

// Claims is an empty claims type used for parsing tokens for viewing purposes
type Claims struct{}

func (c *Claims) Valid() error {
	return nil
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

	// Read from stdin if no files were passed in
	if len(cfg.Files) == 0 {
		cfg.Files = append(cfg.Files, "")
	}

	parser := &jwt.Parser{SkipClaimsValidation: true}
	for _, file := range cfg.Files {
		infile := OpenFileLogError(file)
		if infile == nil {
			continue
		}

		scanner := bufio.NewScanner(infile)
		for scanner.Scan() {
			line := scanner.Text()
			token, parts, err := parser.ParseUnverified(line, &jwt.MapClaims{})
			if err != nil {
				fmt.Printf("'%s': %v\n", line, err)
				continue
			}

			if len(parts) > 0 {
				token.Signature = parts[len(parts)-1]
			}
			displayToken := &Token{token}
			fmt.Println(displayToken)
		}
		if scanner.Err() != nil {
			fmt.Printf("error reading %s: %v\n", file, scanner.Err())
		}

		CloseFileLogError(infile)
	}
}

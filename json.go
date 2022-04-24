package multiconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type JsonFile struct {
	File string
}

func NewJsonFile(file string) *JsonFile {
	return &JsonFile{File: file}
}

func (jf *JsonFile) Load(vars interface{}) error {
	bytes, err := ioutil.ReadFile(jf.File)
	if err != nil {
		return fmt.Errorf("%w file '%s': %v", ErrParse, jf.File, err)
	}
	err = json.Unmarshal(bytes, vars)
	if err != nil {
		return fmt.Errorf("%w file '%s': %v", ErrParse, jf.File, err)
	}
	return nil
}

type JsonDirs struct {
	Dirs []string
}

func NewJsonDirs(dirs []string) *JsonDirs {
	return &JsonDirs{Dirs: dirs}
}

func (jd *JsonDirs) Load(vars interface{}) error {
	for _, dir := range jd.Dirs {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Printf("Failed to read directory: %s: %v", dir, err)
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".conf") {
				continue
			}

			jf := &JsonFile{File: filepath.Join(dir, file.Name())}
			err := jf.Load(vars)
			if err != nil {
				return fmt.Errorf("%w file '%s': %v", ErrParse, filepath.Join(dir, file.Name()), err)
			}
		}
	}
	return nil
}

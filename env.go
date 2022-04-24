package multiconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Env struct {
	BaseName string
}

func NewEnv(baseName string) *Env {
	return &Env{BaseName: baseName}
}

func (env *Env) Load(vars interface{}) error {
	return ProcessVars(env.BaseName, reflect.StructField{}, env, vars)
}

func (env *Env) GetSettingName(f reflect.StructField) string {
	name := strings.ToLower(f.Name)
	//opts := ""
	for _, tag := range []string{"env", "json"} {
		tval, ok := f.Tag.Lookup(tag)
		if !ok {
			continue
		}

		if tval == "-" {
			name = ""
		}
		parts := strings.SplitN(tval, ",", 2)
		name = parts[0]
		//	if len(parts) == 2 {
		//		opts = parts[1]
		//	}
		break
	}
	return name
}

func (env *Env) Process(name string, field reflect.StructField, vars reflect.Value) error {
	data, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}

	switch vars.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(data, 0, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vars.OverflowInt(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vars.SetInt(parsed)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(data, 0, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vars.OverflowUint(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vars.SetUint(parsed)

	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vars.OverflowFloat(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vars.SetFloat(parsed)

	case reflect.Complex64, reflect.Complex128:
		parsed, err := strconv.ParseComplex(data, 128)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vars.OverflowComplex(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vars.SetComplex(parsed)

	case reflect.Bool:
		parsed, err := strconv.ParseBool(data)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		vars.SetBool(parsed)

	case reflect.String:
		vars.SetString(data)

	case reflect.Array, reflect.Slice, reflect.Map:
		tmp := reflect.New(vars.Type())
		err := json.Unmarshal([]byte(data), tmp.Interface())
		if err != nil {
			return fmt.Errorf("%w envvar as json [%s]: %v", ErrParse, name, err)
		}
		vars.Set(tmp.Elem())

	default:
		return fmt.Errorf("%w unknown type [%s]: %#v", ErrInvalidArgs, name, vars.Kind())
	}

	return nil
}

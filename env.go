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

/*
TODO: Remove this funciton as it has been replaced with the generalized ProcessVars()
still needs some tweaks to ProcessVars() though, as you can see below also LoadWithProcessor() is already being used
so this LoadVars() is not currently being used...
ProcessVars() still needs a couple tweaks to make it perfect fo ruse with Flags.go flags.go is requiring the StructField
to be passed into ProcesS() from ProcessVars() and its not currently passing in reflect.StructField to Process()
*/
func (env *Env) LoadVars(name string, vars interface{}) error {
	vals, ok := vars.(reflect.Value)
	if !ok {
		vals = reflect.ValueOf(vars)
	}
	if vals.Kind() == reflect.Ptr {
		vals = vals.Elem()
	}
	switch vals.Kind() {
	case reflect.Struct:
		for i := 0; i < vals.NumField(); i++ {
			fieldVal := vals.Field(i)
			f := vals.Type().Field(i)
			setting := env.GetSettingName(f)
			// skip settings that are "-" or dont have a usable name set
			if setting == "" {
				continue
			}
			// skip unexported fields (they have PkgPath set)
			if len(f.PkgPath) != 0 {
				continue
			}
			name := name
			if name != "" {
				name += "_"
			}
			name += setting
			err := env.LoadVars(name, fieldVal)
			if err != nil {
				return err
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		data, ok := os.LookupEnv(name)
		if !ok {
			break
		}
		parsed, err := strconv.ParseInt(data, 0, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vals.OverflowInt(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vals.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		data, ok := os.LookupEnv(name)
		if !ok {
			break
		}
		parsed, err := strconv.ParseUint(data, 0, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vals.OverflowUint(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vals.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		data, ok := os.LookupEnv(name)
		if !ok {
			break
		}
		parsed, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vals.OverflowFloat(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vals.SetFloat(parsed)
	case reflect.Complex64, reflect.Complex128:
		data, ok := os.LookupEnv(name)
		if !ok {
			break
		}
		parsed, err := strconv.ParseComplex(data, 128)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		if vals.OverflowComplex(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, name, data)
		}
		vals.SetComplex(parsed)
	case reflect.Bool:
		data, ok := os.LookupEnv(name)
		if !ok {
			break
		}
		parsed, err := strconv.ParseBool(data)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, name, err)
		}
		vals.SetBool(parsed)
	case reflect.String:
		data, ok := os.LookupEnv(name)
		if !ok {
			break
		}
		vals.SetString(data)
	case reflect.Array, reflect.Slice, reflect.Map:
		data, ok := os.LookupEnv(name)
		if !ok {
			break
		}
		tmp := reflect.New(vals.Type())
		err := json.Unmarshal([]byte(data), tmp.Interface())
		if err != nil {
			return fmt.Errorf("%w envvar as json [%s]: %v", ErrParse, name, err)
		}
		vals.Set(tmp.Elem())
	default:
		return fmt.Errorf("%w unknown type [%s]: %#v", ErrInvalidArgs, name, vals.Kind())
	}

	return nil
}

func (env *Env) Load(vars interface{}) error {
	//	return env.LoadVars(env.BaseName, vars)
	return env.LoadWithProcessor(vars)
}

func (env *Env) LoadWithProcessor(vars interface{}) error {
	return ProcessVars(env.BaseName, reflect.StructField{}, env, vars)
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

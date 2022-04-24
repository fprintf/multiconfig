package multiconfig

import (
	"fmt"
	"reflect"
	"strings"
)

type Loader interface {
	Load(vars interface{}) error
}

type Processor interface {
	Process(string, reflect.StructField, reflect.Value) error
}

var (
	ErrInvalidArgs = fmt.Errorf("invalid arguments")
	ErrParse       = fmt.Errorf("failed to parse")
	ErrOverflow    = fmt.Errorf("overflow")
)

func getSettingName(f reflect.StructField) string {
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

func ProcessVars(name string, field reflect.StructField, p Processor, vars interface{}) error {
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

			// set the name of the setting or use default formatter
			setting := getSettingName(f)
			if p, ok := p.(interface {
				GetSettingName(reflect.StructField) string
			}); ok {
				setting = p.GetSettingName(f)
			}

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
			err := ProcessVars(name, f, p, fieldVal)
			if err != nil {
				return err
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.Bool,
		reflect.String,
		reflect.Array, reflect.Slice, reflect.Map:
		err := p.Process(name, field, vals)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w unknown type [%s]: %#v", ErrInvalidArgs, name, vals.Kind())
	}

	return nil
}

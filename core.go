package multiconfig

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type MultiLoader interface {
	Get(*ProcessorParams) string
}

func LoadItem(l MultiLoader, params *ProcessorParams) error {
	data := l.Get(params)
	vars := params.vals
	switch vars.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(data, 0, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, params.name, err)
		}
		if vars.OverflowInt(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, params.name, data)
		}
		vars.SetInt(parsed)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(data, 0, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, params.name, err)
		}
		if vars.OverflowUint(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, params.name, data)
		}
		vars.SetUint(parsed)

	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, params.name, err)
		}
		if vars.OverflowFloat(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, params.name, data)
		}
		vars.SetFloat(parsed)

	case reflect.Complex64, reflect.Complex128:
		parsed, err := strconv.ParseComplex(data, 128)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, params.name, err)
		}
		if vars.OverflowComplex(parsed) {
			return fmt.Errorf("%w parsing envvar [%s]: '%s' overflows", ErrOverflow, params.name, data)
		}
		vars.SetComplex(parsed)

	case reflect.Bool:
		parsed, err := strconv.ParseBool(data)
		if err != nil {
			return fmt.Errorf("%w envvar [%s]: %v", ErrParse, params.name, err)
		}
		vars.SetBool(parsed)

	case reflect.String:
		vars.SetString(data)

	case reflect.Array, reflect.Slice, reflect.Map:
		tmp := reflect.New(vars.Type())
		err := json.Unmarshal([]byte(data), tmp.Interface())
		if err != nil {
			return fmt.Errorf("%w envvar as json [%s]: %v", ErrParse, params.name, err)
		}
		vars.Set(tmp.Elem())

	default:
		return fmt.Errorf("%w unknown type [%s]: %#v", ErrInvalidArgs, params.name, vars.Kind())
	}

	return nil
}

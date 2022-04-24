package multiconfig

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

type Flags struct {
	flagSet *flag.FlagSet
	args    map[string]interface{}
}

func NewFlags() *Flags {
	return &Flags{}
}

func getFlagUsage(f reflect.StructField) string {
	flagUsage := f.Tag.Get("usage")
	if flagUsage == "" {
		flagUsage = "undocumented option"
	}
	return flagUsage
}

func (fl *Flags) GetSettingName(f reflect.StructField) string {
	flagName := strings.ToLower(f.Name)
	for _, name := range []string{"arg", "json"} {
		tmp := f.Tag.Get(name)
		if tmp != "" {
			flagName = tmp
			break
		}
	}
	return flagName
}

func (fl *Flags) Process(name string, field reflect.StructField, vars reflect.Value) error {
	flagUsage := getFlagUsage(field)
	switch vars.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		fl.args[field.Name] = fl.flagSet.Int(name, int(vars.Int()), flagUsage)
	case reflect.Int64:
		fl.args[field.Name] = fl.flagSet.Int64(name, vars.Int(), flagUsage)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		fl.args[field.Name] = fl.flagSet.Uint(name, uint(vars.Uint()), flagUsage)
	case reflect.Uint64:
		fl.args[field.Name] = fl.flagSet.Uint64(name, vars.Uint(), flagUsage)
	case reflect.Float64, reflect.Float32:
		fl.args[field.Name] = fl.flagSet.Float64(name, vars.Float(), flagUsage)
	case reflect.Bool:
		fl.args[field.Name] = fl.flagSet.Bool(name, vars.Bool(), flagUsage)
	case reflect.String:
		fl.args[field.Name] = fl.flagSet.String(name, vars.String(), flagUsage)
	case reflect.Array, reflect.Slice, reflect.Map:
		fl.args[field.Name] = fl.flagSet.String(name, vars.String(), flagUsage)
	default:
		return fmt.Errorf("%w unknown type [%s]: %v", ErrInvalidArgs, name, vars.Kind())
	}

	return nil
}

func (fl *Flags) Load(vars interface{}) error {

	fl.flagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.args = make(map[string]interface{})

	vals := reflect.ValueOf(vars)
	if vals.Kind() == reflect.Ptr {
		vals = vals.Elem()
	}

	err := ProcessVars("", reflect.StructField{}, fl, vars)
	if err != nil {
		return err
	}
	// Parse() will ignore parameters unless we skip the program name with os.Args[1:]
	err = fl.flagSet.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("flag parse error: %v", err)
	}

	// TODO decide if we should fail on additional args remaining after parse or not (prob should make it optional)
	for i, arg := range fl.flagSet.Args() {
		log.Printf("remaining arg: %-02d %s", i, arg)
	}

	for i := 0; i < vals.NumField(); i++ {
		fieldVal := vals.Field(i)
		f := vals.Type().Field(i)
		switch fieldVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:

			data, ok := fl.args[f.Name]
			if ok {
				val := reflect.ValueOf(data)
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}

				log.Printf("setting type %T %#v", fieldVal, fieldVal)

				fieldVal.Set(val)
			}

		case reflect.Array, reflect.Slice, reflect.Map:
			data, ok := fl.args[f.Name].(string)
			if !ok {
				break
			}
			tmp := reflect.New(fieldVal.Type())
			err := json.Unmarshal([]byte(data), tmp.Interface())
			if err != nil {
				return err
			}
			fieldVal.Set(reflect.ValueOf(tmp).Elem())

		default:
			// ignore unknown fields
		}
	}

	return nil
}
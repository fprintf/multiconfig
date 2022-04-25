package multiconfig

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
)

type Flags struct {
	flagSet *flag.FlagSet
	args    map[string]interface{}
	OsArgs  []string
}

func NewFlags() *Flags {
	return &Flags{OsArgs: os.Args[1:]}
}
func NewFlagsWithArgs(args []string) *Flags {
	return &Flags{OsArgs: args}
}

func getFlagUsage(f reflect.StructField) string {
	flagUsage := f.Tag.Get("usage")
	if flagUsage == "" {
		flagUsage = "undocumented option"
	}
	return flagUsage
}

func getIndex(index []int) string {
	data, _ := json.Marshal(index)
	return string(data)
}
func getRealIndex(index string) []int {
	realIndex := []int{}
	err := json.Unmarshal([]byte(index), &realIndex)
	if err != nil {
		log.Printf("error getting index: %v", err)
	}
	return realIndex
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

func (fl *Flags) Process(params *ProcessorParams) error {
	flagUsage := getFlagUsage(params.field)
	flagIndex := getIndex(params.index)

	vars := params.vals
	switch vars.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		fl.args[flagIndex] = fl.flagSet.Int(params.name, int(vars.Int()), flagUsage)
	case reflect.Int64:
		fl.args[flagIndex] = fl.flagSet.Int64(params.name, vars.Int(), flagUsage)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		fl.args[flagIndex] = fl.flagSet.Uint(params.name, uint(vars.Uint()), flagUsage)
	case reflect.Uint64:
		fl.args[flagIndex] = fl.flagSet.Uint64(params.name, vars.Uint(), flagUsage)
	case reflect.Float64, reflect.Float32:
		fl.args[flagIndex] = fl.flagSet.Float64(params.name, vars.Float(), flagUsage)
	case reflect.Bool:
		fl.args[flagIndex] = fl.flagSet.Bool(params.name, vars.Bool(), flagUsage)
	case reflect.String:
		if _, ok := vars.Interface().(string); ok && params.field.Tag.Get("argtype") == "positional" {
			fl.args[flagIndex] = ""
			break
		}
		fl.args[flagIndex] = fl.flagSet.String(params.name, vars.String(), flagUsage)
	case reflect.Array, reflect.Slice, reflect.Map:
		if _, ok := vars.Interface().([]string); ok && params.field.Tag.Get("argtype") == "positional" {
			fl.args[flagIndex] = []string{}
			break
		}
		out, _ := json.Marshal(vars.Interface())
		fl.args[flagIndex] = fl.flagSet.String(params.name, string(out), flagUsage)
	default:
		return fmt.Errorf("%w unknown type [%s]: %v", ErrInvalidArgs, params.name, vars.Kind())
	}

	return nil
}

func (fl *Flags) Load(vars interface{}) error {
	fl.flagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.args = make(map[string]interface{})

	err := ProcessVars(ProcessorParams{}, fl, vars)
	if err != nil {
		return err
	}
	// Parse() will ignore parameters unless we skip the program name with os.Args[1:]
	err = fl.flagSet.Parse(fl.OsArgs)
	if err != nil {
		return fmt.Errorf("flag parse error: %v", err)
	}

	positional := fl.flagSet.Args()
	// Set values after getting them from the flag results
	vals := reflect.ValueOf(vars)
	if vals.Kind() == reflect.Ptr {
		vals = vals.Elem()
	}

	// Sort our key list so we add positional arguments in order deterministicly
	sorted := make([]string, 0, len(fl.args))
	for key := range fl.args {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)

	for _, index := range sorted {
		data := fl.args[index]
		realIndex := getRealIndex(index)
		fieldVal := vals.FieldByIndex(realIndex)
		argType := vals.Type().FieldByIndex(realIndex).Tag.Get("argtype")

		switch fieldVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			fieldVal.SetInt(int64(*data.(*int)))
		case reflect.Int64:
			fieldVal.SetInt(*data.(*int64))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			fieldVal.SetUint(uint64(*data.(*uint)))
		case reflect.Uint64:
			fieldVal.SetUint(*data.(*uint64))

		case reflect.Float32, reflect.Float64:
			fieldVal.SetFloat(*data.(*float64))

		case reflect.Bool, reflect.String:
			// Store positional arg and remove it from positional arg list
			if fieldVal.Kind() == reflect.String && argType == "positional" && len(positional) > 0 {
				val := positional[0]
				positional = positional[1:]
				fieldVal.Set(reflect.ValueOf(val))
				break
			}
			val := reflect.ValueOf(data)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			fieldVal.Set(val)

		case reflect.Array, reflect.Slice, reflect.Map:
			// Store remaining positional args and remove it from positional arg list
			if _, ok := fieldVal.Interface().([]string); ok && argType == "positional" {
				fieldVal.Set(reflect.ValueOf(positional))
				positional = []string{}
				break
			}
			data, ok := data.(*string)
			if !ok {
				break
			}
			tmp := reflect.New(fieldVal.Type())
			err := json.Unmarshal([]byte(*data), tmp.Interface())
			if err != nil {
				return fmt.Errorf("failed to parse: [%s]: %v", *data, err)
			}
			fieldVal.Set(tmp.Elem())

		default:
			// ignore unknown fields
		}
	}

	return nil
}

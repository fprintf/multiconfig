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
		fl.args[flagIndex] = fl.flagSet.String(params.name, vars.String(), flagUsage)
	case reflect.Array, reflect.Slice, reflect.Map:
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
	err = fl.flagSet.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("flag parse error: %v", err)
	}
	// TODO decide if we should fail on additional args remaining after parse or not (prob should make it optional)
	for i, arg := range fl.flagSet.Args() {
		log.Printf("remaining arg: %-02d %s", i, arg)
	}

	// Set values after getting them from the flag results
	vals := reflect.ValueOf(vars)
	if vals.Kind() == reflect.Ptr {
		vals = vals.Elem()
	}
	for index, data := range fl.args {
		realIndex := getRealIndex(index)
		fieldVal := vals.FieldByIndex(realIndex)

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
			val := reflect.ValueOf(data)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			fieldVal.Set(val)

		case reflect.Array, reflect.Slice, reflect.Map:
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

// Package env populates Go structs from environment variables using struct
// field tags and provides convenience helpers for reading and writing
// individual environment variables.
//
// Struct tags use the key "env" to map fields to environment variable names:
//
//	type Config struct {
//	    Host string `env:"APP_HOST"`
//	    Port int    `env:"APP_PORT"`
//	}
//
// Nested structs are walked recursively. Fields tagged with "-" are skipped.
// Slices are populated from comma-separated values. Types implementing
// encoding.TextUnmarshaler are handled via their UnmarshalText method.
package env

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	internalTag             = "env"
	internalTagDefaultValue = "default"
	internalTagExclude      = "-"
)

func Unmarshal(v any) error {
	return deserializeStruct(reflect.ValueOf(v))
}

func deserializeValue(ref reflect.Value, value string) error {
	refType := ref.Type()
	if refType == reflect.TypeFor[time.Duration]() {
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		ref.SetInt(int64(dur))
		return nil
	}

	pointer := ref
	if ref.Kind() != reflect.Ptr {
		pointer = ref.Addr()
	}

	if pointer.Type().Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
		m := pointer.MethodByName("UnmarshalText")
		result := m.Call([]reflect.Value{reflect.ValueOf([]byte(value))})
		if !result[0].IsNil() {
			return result[0].Interface().(error)
		}
		return nil
	}

	var aggError error
	switch ref.Kind() {
	case reflect.String:
		ref.SetString(value)
	case reflect.Int:
		num, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return err
		}
		ref.SetInt(num)
	case reflect.Int64:
		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		ref.SetInt(num)
	case reflect.Int8:
		num, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return err
		}
		ref.SetInt(num)
	case reflect.Int16:
		num, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return err
		}
		ref.SetInt(num)
	case reflect.Int32:
		num, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		ref.SetInt(num)
	case reflect.Uint, reflect.Uint64:
		num, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		ref.SetUint(num)
	case reflect.Uint8:
		num, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return err
		}
		ref.SetUint(num)
	case reflect.Uint16:
		num, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		ref.SetUint(num)
	case reflect.Uint32:
		num, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		ref.SetUint(num)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		ref.SetBool(b)
	case reflect.Float64:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		ref.SetFloat(num)
	case reflect.Float32:
		num, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		ref.SetFloat(num)
	case reflect.Slice:
		elemType := refType.Elem()

		if elemType.Kind() == reflect.Slice || elemType.Kind() == reflect.Array {
			ref.SetZero()
			return fmt.Errorf("unsupported type: dimensional array")
		}

		values := strings.Split(value, ",")

		slice := reflect.MakeSlice(ref.Type(), len(values), len(values))

		for index, item := range values {
			err := deserializeValue(slice.Index(index), item)
			if err != nil {
				aggError = errors.Join(aggError, err)
			}
		}
		ref.Set(slice)
	default:
		ref.SetZero()
	}
	return aggError
}

func deserializeStruct(ref reflect.Value) (err error) {
	if ref.Kind() == reflect.Pointer {
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		return fmt.Errorf("failed to deserialize struct: provided %s", ref.Kind())
	}

	refType := ref.Type()
	for index := 0; index < ref.NumField(); index++ {
		field := refType.Field(index)

		tag := field.Tag.Get(internalTag)

		if tag == internalTagExclude {
			continue
		}

		fieldRef := ref.Field(index)

		if !fieldRef.CanSet() {
			continue
		}

		isTextUnmarshaler := fieldRef.Addr().Type().Implements(reflect.TypeFor[encoding.TextUnmarshaler]())
		if fieldRef.Kind() == reflect.Struct && !isTextUnmarshaler {
			err = errors.Join(err, deserializeStruct(fieldRef))
		} else if fieldRef.Kind() == reflect.Pointer && fieldRef.Type().Elem().Kind() == reflect.Struct && !fieldRef.Type().Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
			fieldRef.Set(reflect.New(fieldRef.Type().Elem()))
			err = errors.Join(err, deserializeStruct(fieldRef.Elem()))
		} else {
			value, exists := os.LookupEnv(tag)

			if exists {
				err = errors.Join(err, deserializeValue(fieldRef, value))
			} else {
				defaultValue := field.Tag.Get(internalTagDefaultValue)
				if defaultValue != "" {
					err = errors.Join(err, deserializeValue(fieldRef, defaultValue))
				}
			}

		}
	}
	return
}

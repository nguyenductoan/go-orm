package helpers

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/jackc/pgtype"
)

// Populate dynamic assign value for the reflect.Value
func Populate(v reflect.Value, value interface{}) error {
	switch v.Kind() {
	case reflect.String:
		if _, ok := value.(string); ok {
			v.SetString(value.(string))
		} else {
			v.SetString("invalid")
		}
	case reflect.Slice:
		arr := parseArr(value)
		for _, elem := range arr {
			v.Set(reflect.Append(v, elem))
		}
	case reflect.Array:
		v.Set(reflect.ValueOf(value))
	case reflect.Int:
		i, err := strconv.ParseInt(fmt.Sprintf("%d", value), 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Bool:
		v.SetBool(value.(bool))
	case reflect.Struct:
		v.Set(reflect.ValueOf(value))
	case reflect.Ptr:
	default:
		return fmt.Errorf("unsupported kind %s", v.Type())
	}
	return nil
}

func parseArr(arrValue interface{}) (result []reflect.Value) {
	val := reflect.ValueOf(arrValue)
	switch val.Kind() {
	case reflect.Struct:
		switch arrValue.(type) {
		case pgtype.TextArray:
			arr := arrValue.(pgtype.TextArray)
			elems := arr.Elements
			for _, elem := range elems {
				v, err := elem.Value()
				if v != nil && err == nil {
					result = append(result, reflect.ValueOf(v))
				}
			}
		default:
			fmt.Printf("WARN: type %T is not supported\n", arrValue)
		}
	case reflect.Ptr:
		switch arrValue.(type) {
		case *pgtype.TextArray:
			arr := arrValue.(*pgtype.TextArray)
			elems := arr.Elements
			for _, elem := range elems {
				v, err := elem.Value()
				if v != nil && err == nil {
					result = append(result, reflect.ValueOf(v))
				}
			}
		default:
			fmt.Printf("WARN: type %T is not supported\n", arrValue)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			result = append(result, val.Index(i))
		}
	default:
		fmt.Printf("WARN: kind %T is not supported\n", arrValue)
	}
	return
}

// UnderlyingValue return the interface{} holding concrete value of a reflect.Value
func UnderlyingValue(val reflect.Value) (result interface{}) {
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		return val.Interface()
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Struct:
		return val.Interface()
	default:
		panic(errors.New("convert reflectValue error"))
	}
}

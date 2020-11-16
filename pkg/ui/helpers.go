package ui


import (
	"fmt"
	"strings"
	"reflect"
	"strconv"
)

func interfaceToArrayOfInterface(a interface{}) ([]interface{}, error) {
	object := reflect.ValueOf(a)

	if object.Kind() != reflect.Slice && object.Kind() != reflect.Array {
		return nil, fmt.Errorf("Can't convert not array value into []interface{}")
	}
	items := make([]interface{}, object.Len())
	for i := 0; i < object.Len(); i++ {
		items[i] = object.Index(i).Interface()
	}
	return items, nil
}

// UnwrapValuePtr recursively unwrap pointers
func UnwrapValuePtr(v reflect.Value) *reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return UnwrapValuePtr(v.Elem())
	}
	return &v
}

// UnwrapTypePtr recursively unwrap pointers
func UnwrapTypePtr(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return UnwrapTypePtr(t.Elem())
	}
	return t
}

// StringifyValue stringify any reflect.Value
func StringifyValue(ft reflect.Value) string {
	switch ft.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(ft.Int(), 10)
	case reflect.String:
		return ft.String()
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", ft.Float())
	case reflect.Ptr:
		if ft.IsNil() {
			return ""
		}
		return StringifyValue(ft.Elem())
	default:
		return string(ft.String())
	}
}


func getNesstedVal(v reflect.Value, path []string) (val *reflect.Value) {
	val = &v

	for _, key := range path {
		// unwrap the pointers
		val = UnwrapValuePtr(val.FieldByName(key))
		if val == nil {
			return
		}
	}
	return 
}

func getNesstedType(t reflect.Type, path []string) (*reflect.Type) {
	nestedType := t

	for _, key := range path {
		structField, found := nestedType.FieldByName(key)
		if !found {
			return nil
		}
		// unwrap the pointers
		nestedType = UnwrapTypePtr(structField.Type)
	}
	return &nestedType
}

func EnsureStringPaths(obj interface{}, paths []string) []string {
	objType := reflect.TypeOf(obj)
	for _, path := range paths {
		if getNesstedType(objType, strings.Split(path, ".")) == nil {
			panic(fmt.Sprintf("[EnsureStringPaths]:: Not found path: '%s' on type: %s",path ,objType.Name()))
		}
	}
	return paths
}

func Contains(s []string, searchterm string) bool {
	for _, a := range s {
        if a ==  searchterm {
            return true
        }
    }
    return false
}


func multiStr(s string, len int) string {
	str := ""
	for i :=0; i<len; i++ {
		str += s
	}
	return str
}
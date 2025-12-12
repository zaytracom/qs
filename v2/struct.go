// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ParseToStruct parses a query string and fills a struct using query tags.
//
// The function parses the query string into a map using Parse(), then maps
// the values to struct fields based on their `query` tags.
//
// Example:
//
//	type User struct {
//	    Name  string `query:"name"`
//	    Age   int    `query:"age"`
//	    Email string `query:"email"`
//	}
//
//	var user User
//	err := qs.ParseToStruct("name=John&age=30&email=john@example.com", &user)
//	// user = User{Name: "John", Age: 30, Email: "john@example.com"}
//
// Supported field types:
//   - string
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - bool
//   - time.Time (parsed from RFC3339 format)
//   - []T (slices of any supported type)
//   - map[string]T (maps with string keys)
//   - nested structs
//   - pointers to any supported type
func ParseToStruct(str string, dest any, opts ...ParseOption) error {
	// Parse to map first
	result, err := Parse(str, opts...)
	if err != nil {
		return err
	}

	// Convert map to struct
	return MapToStruct(result, dest)
}

// MapToStruct converts a map[string]any to a struct using query tags.
//
// Fields are matched by their `query` tag. If no tag is present, the lowercase
// field name is used. Use `query:"-"` to skip a field.
//
// Example:
//
//	data := map[string]any{"name": "John", "age": "30"}
//	var user User
//	err := qs.MapToStruct(data, &user)
func MapToStruct(data map[string]any, dest any) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer to struct, got %T", dest)
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to struct, got pointer to %s", destValue.Kind())
	}

	return fillStruct(data, destValue)
}

// StructToQueryString converts a struct to a query string using query tags.
//
// Example:
//
//	user := User{Name: "John", Age: 30, Email: "john@example.com"}
//	str, err := qs.StructToQueryString(user)
//	// str = "age=30&email=john%40example.com&name=John"
func StructToQueryString(obj any, opts ...StringifyOption) (string, error) {
	data, err := StructToMap(obj)
	if err != nil {
		return "", err
	}

	return Stringify(data, opts...)
}

// StructToMap converts a struct to a map[string]any using query tags.
//
// Fields are named by their `query` tag. If no tag is present, the lowercase
// field name is used. Use `query:"-"` to skip a field.
func StructToMap(obj any) (map[string]any, error) {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		if objValue.IsNil() {
			return nil, nil
		}
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object must be a struct or pointer to struct, got %s", objValue.Kind())
	}

	return marshalStruct(objValue)
}

// Unmarshal parses a query string and stores the result in the value pointed to by v.
//
// This function provides idiomatic Go unmarshaling with automatic type detection.
// It works with structs, maps, slices, and primitive types.
//
// Example:
//
//	// Unmarshal to struct
//	var user User
//	err := qs.Unmarshal("name=John&age=30", &user)
//
//	// Unmarshal to map
//	var data map[string]any
//	err := qs.Unmarshal("name=John&age=30", &data)
//
//	// Unmarshal to interface{}
//	var result any
//	err := qs.Unmarshal("name=John&age=30", &result)
func Unmarshal(queryString string, v any, opts ...ParseOption) error {
	if v == nil {
		return fmt.Errorf("unmarshal target cannot be nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("unmarshal target must be a pointer, got %T", v)
	}

	rv = rv.Elem()
	if !rv.CanSet() {
		return fmt.Errorf("unmarshal target must be settable")
	}

	// Parse the query string to map first
	data, err := Parse(queryString, opts...)
	if err != nil {
		return err
	}

	return unmarshalValue(data, rv)
}

// Marshal converts a value to a query string.
//
// This function provides idiomatic Go marshaling with automatic type detection.
// It works with structs, maps, slices, and primitive types.
//
// Example:
//
//	// Marshal struct
//	user := User{Name: "John", Age: 30}
//	str, err := qs.Marshal(user)
//	// str = "age=30&name=John"
//
//	// Marshal map
//	data := map[string]any{"name": "John", "age": 30}
//	str, err := qs.Marshal(data)
func Marshal(v any, opts ...StringifyOption) (string, error) {
	if v == nil {
		return "", nil
	}

	data, err := marshalValue(v)
	if err != nil {
		return "", err
	}

	// If marshaled value is nil or not a map, return empty string
	dataMap, ok := data.(map[string]any)
	if !ok || dataMap == nil {
		return "", nil
	}

	return Stringify(dataMap, opts...)
}

// fillStruct recursively fills struct fields from map data.
func fillStruct(data map[string]any, structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get query tag
		queryTag := getQueryTag(fieldType)
		if queryTag == "-" {
			continue
		}

		// Look for the value in data
		value, exists := data[queryTag]
		if !exists {
			continue
		}

		if err := setFieldValue(field, value); err != nil {
			return fmt.Errorf("error setting field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// getQueryTag returns the query tag name for a struct field.
// Falls back to lowercase field name if no tag is present.
func getQueryTag(field reflect.StructField) string {
	tag := field.Tag.Get("query")
	if tag == "" {
		return strings.ToLower(field.Name)
	}
	// Handle comma-separated options (e.g., `query:"name,omitempty"`)
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}

// setFieldValue sets a struct field value from any data.
func setFieldValue(field reflect.Value, value any) error {
	if value == nil {
		return nil
	}

	fieldType := field.Type()

	// Handle pointers
	if fieldType.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		return setFieldValue(field.Elem(), value)
	}

	// Handle time.Time specially
	if fieldType == reflect.TypeOf(time.Time{}) {
		return setTimeField(field, value)
	}

	valueReflect := reflect.ValueOf(value)

	switch fieldType.Kind() {
	case reflect.String:
		return setStringField(field, value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setIntField(field, value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUintField(field, value)

	case reflect.Float32, reflect.Float64:
		return setFloatField(field, value)

	case reflect.Bool:
		return setBoolField(field, value)

	case reflect.Slice:
		return setSliceField(field, value)

	case reflect.Struct:
		if dataMap, ok := value.(map[string]any); ok {
			return fillStruct(dataMap, field)
		}
		return fmt.Errorf("cannot convert %T to struct", value)

	case reflect.Map:
		if fieldType.Key().Kind() == reflect.String {
			return setMapField(field, value)
		}
		return fmt.Errorf("unsupported map key type: %v", fieldType.Key().Kind())

	case reflect.Interface:
		// For interface{}, set the value directly
		field.Set(valueReflect)
		return nil

	default:
		// Try direct assignment if types match
		if valueReflect.Type().AssignableTo(fieldType) {
			field.Set(valueReflect)
			return nil
		}
		return fmt.Errorf("unsupported field type: %v", fieldType.Kind())
	}
}

// setStringField sets a string field from any value.
func setStringField(field reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		field.SetString(v)
	default:
		field.SetString(fmt.Sprintf("%v", value))
	}
	return nil
}

// setIntField sets an int field from any value.
func setIntField(field reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		intVal, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to int: %w", v, err)
		}
		field.SetInt(intVal)
	case int:
		field.SetInt(int64(v))
	case int64:
		field.SetInt(v)
	case float64:
		field.SetInt(int64(v))
	default:
		return fmt.Errorf("cannot convert %T to int", value)
	}
	return nil
}

// setUintField sets a uint field from any value.
func setUintField(field reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		uintVal, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to uint: %w", v, err)
		}
		field.SetUint(uintVal)
	case uint:
		field.SetUint(uint64(v))
	case uint64:
		field.SetUint(v)
	case int:
		if v < 0 {
			return fmt.Errorf("cannot convert negative int to uint")
		}
		field.SetUint(uint64(v))
	case int64:
		if v < 0 {
			return fmt.Errorf("cannot convert negative int64 to uint")
		}
		field.SetUint(uint64(v))
	case float64:
		if v < 0 {
			return fmt.Errorf("cannot convert negative float to uint")
		}
		field.SetUint(uint64(v))
	default:
		return fmt.Errorf("cannot convert %T to uint", value)
	}
	return nil
}

// setFloatField sets a float field from any value.
func setFloatField(field reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		floatVal, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to float: %w", v, err)
		}
		field.SetFloat(floatVal)
	case float64:
		field.SetFloat(v)
	case float32:
		field.SetFloat(float64(v))
	case int:
		field.SetFloat(float64(v))
	case int64:
		field.SetFloat(float64(v))
	default:
		return fmt.Errorf("cannot convert %T to float", value)
	}
	return nil
}

// setBoolField sets a bool field from any value.
func setBoolField(field reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		boolVal, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("cannot convert %q to bool: %w", v, err)
		}
		field.SetBool(boolVal)
	case bool:
		field.SetBool(v)
	default:
		return fmt.Errorf("cannot convert %T to bool", value)
	}
	return nil
}

// setTimeField sets a time.Time field from any value.
func setTimeField(field reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		// Try RFC3339 first
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Try RFC3339Nano
			t, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				// Try ISO8601 without timezone
				t, err = time.Parse("2006-01-02T15:04:05", v)
				if err != nil {
					// Try date only
					t, err = time.Parse("2006-01-02", v)
					if err != nil {
						return fmt.Errorf("cannot parse time %q: %w", v, err)
					}
				}
			}
		}
		field.Set(reflect.ValueOf(t))
	case time.Time:
		field.Set(reflect.ValueOf(v))
	default:
		return fmt.Errorf("cannot convert %T to time.Time", value)
	}
	return nil
}

// setSliceField sets a slice field from any value.
func setSliceField(field reflect.Value, value any) error {
	var sliceValue []any

	switch v := value.(type) {
	case []any:
		sliceValue = v
	case map[string]any:
		// Convert map with numeric keys to slice
		sliceValue = mapToSlice(v)
	default:
		// Single value becomes slice with one element
		sliceValue = []any{value}
	}

	fieldType := field.Type()
	newSlice := reflect.MakeSlice(fieldType, len(sliceValue), len(sliceValue))

	for i, item := range sliceValue {
		elemField := newSlice.Index(i)
		if err := setFieldValue(elemField, item); err != nil {
			return fmt.Errorf("error setting slice element %d: %w", i, err)
		}
	}

	field.Set(newSlice)
	return nil
}

// setMapField sets a map field from any value.
func setMapField(field reflect.Value, value any) error {
	dataMap, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("cannot convert %T to map", value)
	}

	fieldType := field.Type()
	valueType := fieldType.Elem()

	newMap := reflect.MakeMap(fieldType)

	for k, v := range dataMap {
		keyVal := reflect.ValueOf(k)
		valueVal := reflect.New(valueType).Elem()

		if err := setFieldValue(valueVal, v); err != nil {
			return fmt.Errorf("error setting map value for key %q: %w", k, err)
		}

		newMap.SetMapIndex(keyVal, valueVal)
	}

	field.Set(newMap)
	return nil
}

// mapToSlice converts a map with numeric string keys to a slice.
func mapToSlice(m map[string]any) []any {
	if len(m) == 0 {
		return []any{}
	}

	// Find the maximum index
	maxIndex := -1
	for k := range m {
		idx, err := strconv.Atoi(k)
		if err == nil && idx > maxIndex {
			maxIndex = idx
		}
	}

	if maxIndex < 0 {
		// No numeric keys, can't convert
		return []any{m}
	}

	result := make([]any, maxIndex+1)
	for k, v := range m {
		idx, err := strconv.Atoi(k)
		if err == nil && idx >= 0 && idx <= maxIndex {
			result[idx] = v
		}
	}

	return result
}

// unmarshalValue recursively unmarshals data into a reflect.Value.
func unmarshalValue(data any, rv reflect.Value) error {
	if data == nil {
		return nil
	}

	rt := rv.Type()

	// Handle pointers
	if rt.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rt.Elem()))
		}
		return unmarshalValue(data, rv.Elem())
	}

	switch rt.Kind() {
	case reflect.Struct:
		if dataMap, ok := data.(map[string]any); ok {
			return fillStruct(dataMap, rv)
		}
		return fmt.Errorf("cannot unmarshal %T into struct", data)

	case reflect.Map:
		return unmarshalMap(data, rv)

	case reflect.Slice:
		return unmarshalSlice(data, rv)

	case reflect.Interface:
		// For interface{}, set the data directly
		if rt == reflect.TypeOf((*any)(nil)).Elem() {
			rv.Set(reflect.ValueOf(data))
			return nil
		}
		return fmt.Errorf("unsupported interface type: %v", rt)

	default:
		// Handle primitive types
		return setFieldValue(rv, data)
	}
}

// unmarshalMap unmarshals data into a map.
func unmarshalMap(data any, rv reflect.Value) error {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("cannot unmarshal %T into map", data)
	}

	rt := rv.Type()
	keyType := rt.Key()
	valueType := rt.Elem()

	// Only support string keys for now
	if keyType.Kind() != reflect.String {
		return fmt.Errorf("unsupported map key type: %v", keyType)
	}

	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rt))
	}

	for k, v := range dataMap {
		keyVal := reflect.ValueOf(k)
		valueVal := reflect.New(valueType).Elem()

		if err := unmarshalValue(v, valueVal); err != nil {
			return fmt.Errorf("error unmarshaling map value for key %q: %w", k, err)
		}

		rv.SetMapIndex(keyVal, valueVal)
	}

	return nil
}

// unmarshalSlice unmarshals data into a slice.
func unmarshalSlice(data any, rv reflect.Value) error {
	var sliceData []any

	switch v := data.(type) {
	case []any:
		sliceData = v
	case map[string]any:
		sliceData = mapToSlice(v)
	default:
		// Single value becomes slice with one element
		sliceData = []any{data}
	}

	rt := rv.Type()
	newSlice := reflect.MakeSlice(rt, len(sliceData), len(sliceData))

	for i, item := range sliceData {
		elemVal := newSlice.Index(i)
		if err := unmarshalValue(item, elemVal); err != nil {
			return fmt.Errorf("error unmarshaling slice element %d: %w", i, err)
		}
	}

	rv.Set(newSlice)
	return nil
}

// marshalValue converts a value to a format suitable for Stringify.
func marshalValue(v any) (any, error) {
	if v == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(v)
	return marshalReflectValue(rv)
}

// marshalReflectValue converts a reflect.Value to a format suitable for Stringify.
func marshalReflectValue(rv reflect.Value) (any, error) {
	// Handle pointers
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		return marshalReflectValue(rv.Elem())
	}

	// Handle time.Time specially
	if rv.Type() == reflect.TypeOf(time.Time{}) {
		t := rv.Interface().(time.Time)
		if t.IsZero() {
			return nil, nil
		}
		return t, nil
	}

	switch rv.Kind() {
	case reflect.Struct:
		return marshalStruct(rv)
	case reflect.Map:
		return marshalMap(rv)
	case reflect.Slice:
		return marshalSlice(rv)
	case reflect.Interface:
		if rv.IsNil() {
			return nil, nil
		}
		return marshalReflectValue(rv.Elem())
	default:
		// Return primitive values as-is
		return rv.Interface(), nil
	}
}

// marshalStruct converts a struct to a map using query tags.
func marshalStruct(rv reflect.Value) (map[string]any, error) {
	result := make(map[string]any)
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get query tag
		queryTag := getQueryTag(fieldType)
		if queryTag == "-" {
			continue
		}

		// Skip nil pointers
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// Skip zero time.Time
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t := field.Interface().(time.Time)
			if t.IsZero() {
				continue
			}
		}

		// Marshal field value
		fieldValue, err := marshalReflectValue(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling field %s: %w", fieldType.Name, err)
		}

		if fieldValue != nil {
			result[queryTag] = fieldValue
		}
	}

	return result, nil
}

// marshalMap converts a map to a format suitable for Stringify.
func marshalMap(rv reflect.Value) (map[string]any, error) {
	if rv.IsNil() {
		return nil, nil
	}

	result := make(map[string]any)

	for _, key := range rv.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		value := rv.MapIndex(key)

		marshaledValue, err := marshalReflectValue(value)
		if err != nil {
			return nil, fmt.Errorf("error marshaling map value for key %q: %w", keyStr, err)
		}

		if marshaledValue != nil {
			result[keyStr] = marshaledValue
		}
	}

	return result, nil
}

// marshalSlice converts a slice to []any.
func marshalSlice(rv reflect.Value) ([]any, error) {
	if rv.IsNil() {
		return nil, nil
	}

	result := make([]any, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		marshaledElem, err := marshalReflectValue(elem)
		if err != nil {
			return nil, fmt.Errorf("error marshaling slice element %d: %w", i, err)
		}
		result[i] = marshaledElem
	}

	return result, nil
}

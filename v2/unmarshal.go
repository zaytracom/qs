// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/zaytracom/qs/v2/lang"
)

// structInfo caches field information for a struct type.
type structInfo struct {
	fields map[string]fieldInfo // query tag → field info
}

// fieldInfo holds information about a single struct field.
type fieldInfo struct {
	index     []int        // field index path for embedded structs
	fieldType reflect.Type // field type
}

// typeCache caches struct information to avoid repeated reflection.
var typeCache sync.Map // map[reflect.Type]*structInfo

// getStructInfo returns cached struct info or builds it.
func getStructInfo(t reflect.Type) *structInfo {
	if cached, ok := typeCache.Load(t); ok {
		return cached.(*structInfo)
	}

	info := buildStructInfo(t)
	typeCache.Store(t, info)
	return info
}

// buildStructInfo builds struct info via reflection.
func buildStructInfo(t reflect.Type) *structInfo {
	info := &structInfo{
		fields: make(map[string]fieldInfo),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get query tag
		tag := field.Tag.Get("query")
		if tag == "-" {
			continue
		}

		// Parse tag options
		name := tag
		if idx := strings.Index(tag, ","); idx != -1 {
			name = tag[:idx]
		}
		if name == "" {
			name = strings.ToLower(field.Name)
		}

		info.fields[name] = fieldInfo{
			index:     field.Index,
			fieldType: field.Type,
		}
	}

	return info
}

// UnmarshalOptions configures the behavior of UnmarshalBytes/UnmarshalString.
type UnmarshalOptions struct {
	ParseOptions // embed ParseOptions

	// Reusable arena (optional, for high-performance use).
	Arena *lang.Arena
}

// UnmarshalBytes parses []byte query string directly into dest without intermediate map.
// This is the most efficient way to parse query strings.
//
// dest must be a pointer to:
//   - struct: fields are matched by `query` tag or lowercase field name
//   - map[string]any: parsed as nested map
//   - *any (interface{}): creates map[string]any
//
// Example:
//
//	type User struct {
//	    Name string `query:"name"`
//	    Age  int    `query:"age"`
//	}
//	var user User
//	err := qs.UnmarshalBytes([]byte("name=John&age=30"), &user)
func UnmarshalBytes(data []byte, dest any, opts ...ParseOption) error {
	if dest == nil {
		return fmt.Errorf("unmarshal target cannot be nil")
	}

	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("unmarshal target must be a pointer, got %T", dest)
	}

	rv = rv.Elem()
	if !rv.CanSet() {
		return fmt.Errorf("unmarshal target must be settable")
	}

	// Apply parse options
	options := applyParseOptions(opts...)
	normalizedOpts, err := normalizeParseOptions(&options)
	if err != nil {
		return err
	}

	// Build AST config
	cfg := buildLangConfig(&normalizedOpts)

	// Create or reuse arena
	arena := lang.NewArena(estimateParams(string(data)))

	// Parse AST
	qs, detectedCharset, err := lang.ParseBytes(arena, data, cfg)
	if err != nil {
		if err == lang.ErrParameterLimitExceeded {
			return ErrParameterLimitExceeded
		}
		if err == lang.ErrDepthLimitExceeded {
			return ErrDepthLimitExceeded
		}
		return err
	}

	// Use detected charset
	charset := normalizedOpts.Charset
	if normalizedOpts.CharsetSentinel {
		charset = charsetFromLang(detectedCharset)
	}

	// Create unmarshaler
	u := &unmarshaler{
		arena:   arena,
		qs:      qs,
		charset: charsetToLang(charset),
		opts:    &normalizedOpts,
	}

	return u.unmarshalTo(rv)
}

// Unmarshal parses a query string and stores the result in the value pointed to by dest.
//
// Example:
//
//	var user User
//	err := qs.Unmarshal("name=John&age=30", &user)
func Unmarshal(data string, dest any, opts ...ParseOption) error {
	return UnmarshalBytes([]byte(data), dest, opts...)
}

// unmarshaler holds the state for unmarshaling.
type unmarshaler struct {
	arena   *lang.Arena
	qs      lang.QueryString
	charset lang.Charset
	opts    *ParseOptions
}

// unmarshalTo dispatches to the appropriate unmarshal method based on destination type.
func (u *unmarshaler) unmarshalTo(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Struct:
		return u.unmarshalToStruct(rv)
	case reflect.Map:
		return u.unmarshalToMap(rv)
	case reflect.Interface:
		// dest is *any — create map
		m := make(map[string]any)
		mapVal := reflect.ValueOf(m)
		if err := u.unmarshalToMap(mapVal); err != nil {
			return err
		}
		rv.Set(mapVal)
		return nil
	case reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return u.unmarshalTo(rv.Elem())
	default:
		return fmt.Errorf("unsupported destination type: %v", rv.Kind())
	}
}

// unmarshalToStruct unmarshals AST directly into a struct.
func (u *unmarshaler) unmarshalToStruct(rv reflect.Value) error {
	info := getStructInfo(rv.Type())

	// Group params by root key
	type paramGroup struct {
		rootKey string
		params  []int // indices into arena.Params
	}
	groups := make(map[string]*paramGroup)
	groupOrder := make([]string, 0, u.qs.ParamLen)

	for i := uint16(0); i < u.qs.ParamLen; i++ {
		param := u.arena.Params[i]
		if param.Key.SegLen == 0 {
			continue
		}

		// Get root segment
		rootSeg := u.arena.Segments[param.Key.SegStart]
		rootKey := u.arena.DecodeString(rootSeg.Span, u.charset)

		if g, ok := groups[rootKey]; ok {
			g.params = append(g.params, int(i))
		} else {
			groups[rootKey] = &paramGroup{
				rootKey: rootKey,
				params:  []int{int(i)},
			}
			groupOrder = append(groupOrder, rootKey)
		}
	}

	// Process each group
	for _, rootKey := range groupOrder {
		group := groups[rootKey]

		// Find matching field
		fi, ok := info.fields[rootKey]
		if !ok {
			continue // ignore unknown fields
		}

		field := rv.FieldByIndex(fi.index)
		if !field.CanSet() {
			continue
		}

		// Handle based on field type and param structure
		if err := u.setFieldFromParams(field, group.params); err != nil {
			return fmt.Errorf("error setting field %s: %w", rootKey, err)
		}
	}

	return nil
}

// setFieldFromParams sets a struct field from a group of params.
func (u *unmarshaler) setFieldFromParams(field reflect.Value, paramIndices []int) error {
	if len(paramIndices) == 0 {
		return nil
	}

	fieldType := field.Type()

	// Handle pointers
	if fieldType.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		return u.setFieldFromParams(field.Elem(), paramIndices)
	}

	// Single param with no nested segments → simple value
	if len(paramIndices) == 1 {
		param := u.arena.Params[paramIndices[0]]
		if param.Key.SegLen == 1 {
			return u.setSimpleValue(field, param)
		}
	}

	// Check if this is a nested structure
	switch fieldType.Kind() {
	case reflect.Struct:
		return u.unmarshalNestedStruct(field, paramIndices)
	case reflect.Map:
		return u.unmarshalNestedMap(field, paramIndices)
	case reflect.Slice:
		return u.unmarshalSlice(field, paramIndices)
	default:
		// For simple types, just use the last value
		param := u.arena.Params[paramIndices[len(paramIndices)-1]]
		return u.setSimpleValue(field, param)
	}
}

// setSimpleValue sets a simple (non-nested) field value.
func (u *unmarshaler) setSimpleValue(field reflect.Value, param lang.Param) error {
	val := u.extractValue(param)
	return setFieldValue(field, val)
}

// extractValue extracts value from param.
func (u *unmarshaler) extractValue(param lang.Param) any {
	if !param.HasEquals {
		if u.opts.StrictNullHandling {
			return ExplicitNullValue
		}
		return ""
	}

	if param.ValueIdx == 0xFFFF {
		return ""
	}

	v := u.arena.Values[param.ValueIdx]
	switch v.Kind {
	case lang.ValNull:
		if u.opts.StrictNullHandling {
			return ExplicitNullValue
		}
		return ""
	case lang.ValComma:
		parts := make([]any, v.PartsLen)
		for j := uint8(0); j < v.PartsLen; j++ {
			partSpan := u.arena.ValueParts[int(v.PartsOff)+int(j)]
			parts[j] = u.arena.DecodeString(partSpan, u.charset)
		}
		return parts
	default:
		return u.arena.DecodeString(v.Raw, u.charset)
	}
}

// unmarshalNestedStruct handles nested struct fields.
func (u *unmarshaler) unmarshalNestedStruct(field reflect.Value, paramIndices []int) error {
	info := getStructInfo(field.Type())

	// Group by second segment
	type paramGroup struct {
		key    string
		params []int
	}
	groups := make(map[string]*paramGroup)
	groupOrder := make([]string, 0)

	for _, idx := range paramIndices {
		param := u.arena.Params[idx]
		if param.Key.SegLen < 2 {
			continue // not nested
		}

		seg := u.arena.Segments[param.Key.SegStart+1]
		key := u.arena.DecodeString(seg.Span, u.charset)

		if g, ok := groups[key]; ok {
			g.params = append(g.params, idx)
		} else {
			groups[key] = &paramGroup{key: key, params: []int{idx}}
			groupOrder = append(groupOrder, key)
		}
	}

	for _, key := range groupOrder {
		group := groups[key]

		fi, ok := info.fields[key]
		if !ok {
			continue
		}

		nestedField := field.FieldByIndex(fi.index)
		if !nestedField.CanSet() {
			continue
		}

		// Shift segments and recurse
		if err := u.setNestedFieldFromParams(nestedField, group.params, 1); err != nil {
			return err
		}
	}

	return nil
}

// setNestedFieldFromParams sets a field considering segment depth.
func (u *unmarshaler) setNestedFieldFromParams(field reflect.Value, paramIndices []int, depth int) error {
	if len(paramIndices) == 0 {
		return nil
	}

	fieldType := field.Type()

	// Handle pointers
	if fieldType.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		return u.setNestedFieldFromParams(field.Elem(), paramIndices, depth)
	}

	// Check if all params have only depth+1 segments (leaf values)
	allLeaf := true
	for _, idx := range paramIndices {
		param := u.arena.Params[idx]
		if int(param.Key.SegLen) > depth+1 {
			allLeaf = false
			break
		}
	}

	if allLeaf {
		// Collect values
		switch fieldType.Kind() {
		case reflect.Slice:
			return u.unmarshalSliceAtDepth(field, paramIndices, depth)
		default:
			// Use last value
			param := u.arena.Params[paramIndices[len(paramIndices)-1]]
			return u.setSimpleValue(field, param)
		}
	}

	// Has deeper nesting
	switch fieldType.Kind() {
	case reflect.Struct:
		return u.unmarshalNestedStructAtDepth(field, paramIndices, depth)
	case reflect.Map:
		return u.unmarshalNestedMapAtDepth(field, paramIndices, depth)
	case reflect.Slice:
		return u.unmarshalSliceAtDepth(field, paramIndices, depth)
	default:
		param := u.arena.Params[paramIndices[len(paramIndices)-1]]
		return u.setSimpleValue(field, param)
	}
}

// unmarshalNestedStructAtDepth handles struct at specific depth.
func (u *unmarshaler) unmarshalNestedStructAtDepth(field reflect.Value, paramIndices []int, depth int) error {
	info := getStructInfo(field.Type())

	groups := make(map[string][]int)
	groupOrder := make([]string, 0)

	for _, idx := range paramIndices {
		param := u.arena.Params[idx]
		if int(param.Key.SegLen) <= depth+1 {
			continue
		}

		seg := u.arena.Segments[int(param.Key.SegStart)+depth+1]
		key := u.arena.DecodeString(seg.Span, u.charset)

		if _, ok := groups[key]; !ok {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], idx)
	}

	for _, key := range groupOrder {
		indices := groups[key]

		fi, ok := info.fields[key]
		if !ok {
			continue
		}

		nestedField := field.FieldByIndex(fi.index)
		if !nestedField.CanSet() {
			continue
		}

		if err := u.setNestedFieldFromParams(nestedField, indices, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// unmarshalNestedMap handles map fields at root level.
func (u *unmarshaler) unmarshalNestedMap(field reflect.Value, paramIndices []int) error {
	return u.unmarshalNestedMapAtDepth(field, paramIndices, 0)
}

// unmarshalNestedMapAtDepth handles map at specific depth.
func (u *unmarshaler) unmarshalNestedMapAtDepth(field reflect.Value, paramIndices []int, depth int) error {
	if field.IsNil() {
		field.Set(reflect.MakeMap(field.Type()))
	}

	valueType := field.Type().Elem()

	groups := make(map[string][]int)
	groupOrder := make([]string, 0)

	for _, idx := range paramIndices {
		param := u.arena.Params[idx]
		if int(param.Key.SegLen) <= depth+1 {
			continue
		}

		seg := u.arena.Segments[int(param.Key.SegStart)+depth+1]
		key := u.arena.DecodeString(seg.Span, u.charset)

		if _, ok := groups[key]; !ok {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], idx)
	}

	for _, key := range groupOrder {
		indices := groups[key]

		// Create value for this key
		valueVal := reflect.New(valueType).Elem()

		if err := u.setNestedFieldFromParams(valueVal, indices, depth+1); err != nil {
			return err
		}

		field.SetMapIndex(reflect.ValueOf(key), valueVal)
	}

	return nil
}

// unmarshalSlice handles slice fields at root level.
func (u *unmarshaler) unmarshalSlice(field reflect.Value, paramIndices []int) error {
	return u.unmarshalSliceAtDepth(field, paramIndices, 0)
}

// unmarshalSliceAtDepth handles slice at specific depth.
func (u *unmarshaler) unmarshalSliceAtDepth(field reflect.Value, paramIndices []int, depth int) error {
	elemType := field.Type().Elem()

	// Collect values or indices
	type sliceEntry struct {
		index   int
		params  []int
		isArray bool
	}
	entries := make(map[int]*sliceEntry)
	maxIndex := -1
	hasEmptyBracket := false

	for _, idx := range paramIndices {
		param := u.arena.Params[idx]

		segIdx := depth + 1
		if int(param.Key.SegLen) <= segIdx {
			// Value at this level
			if !hasEmptyBracket {
				hasEmptyBracket = true
				if entries[-1] == nil {
					entries[-1] = &sliceEntry{index: -1, params: nil}
				}
			}
			entries[-1].params = append(entries[-1].params, idx)
			continue
		}

		seg := u.arena.Segments[int(param.Key.SegStart)+segIdx]

		switch seg.Kind {
		case lang.SegEmpty:
			hasEmptyBracket = true
			if entries[-1] == nil {
				entries[-1] = &sliceEntry{index: -1, params: nil}
			}
			entries[-1].params = append(entries[-1].params, idx)
		case lang.SegIndex:
			arrayIdx := int(seg.Index)
			if arrayIdx > maxIndex {
				maxIndex = arrayIdx
			}
			if entries[arrayIdx] == nil {
				entries[arrayIdx] = &sliceEntry{index: arrayIdx, isArray: true}
			}
			entries[arrayIdx].params = append(entries[arrayIdx].params, idx)
		default:
			// Non-numeric key - treat as object, not array
			// For now, append as sequential
			if entries[-1] == nil {
				entries[-1] = &sliceEntry{index: -1, params: nil}
			}
			entries[-1].params = append(entries[-1].params, idx)
		}
	}

	// Build slice
	var result []reflect.Value

	// First add indexed entries
	if maxIndex >= 0 {
		result = make([]reflect.Value, maxIndex+1)
		for i := 0; i <= maxIndex; i++ {
			entry := entries[i]
			elem := reflect.New(elemType).Elem()
			if entry != nil && len(entry.params) > 0 {
				if err := u.setNestedFieldFromParams(elem, entry.params, depth+1); err != nil {
					return err
				}
			}
			result[i] = elem
		}
	}

	// Then add empty bracket entries
	if entry := entries[-1]; entry != nil && len(entry.params) > 0 {
		for _, idx := range entry.params {
			elem := reflect.New(elemType).Elem()
			param := u.arena.Params[idx]
			if err := u.setSimpleValue(elem, param); err != nil {
				return err
			}
			result = append(result, elem)
		}
	}

	// Set the slice
	if len(result) > 0 {
		slice := reflect.MakeSlice(field.Type(), len(result), len(result))
		for i, v := range result {
			slice.Index(i).Set(v)
		}
		field.Set(slice)
	}

	return nil
}

// unmarshalToMap unmarshals AST directly into a map[string]any.
func (u *unmarshaler) unmarshalToMap(rv reflect.Value) error {
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}

	// For map, we need to build the nested structure
	// This is similar to the current Parse but more direct
	result := make(map[string]any)

	for i := uint16(0); i < u.qs.ParamLen; i++ {
		param := u.arena.Params[i]
		if param.Key.SegLen == 0 {
			continue
		}

		// Build key chain
		chain := make([]string, 0, param.Key.SegLen)
		for j := uint8(0); j < param.Key.SegLen; j++ {
			seg := u.arena.Segments[int(param.Key.SegStart)+int(j)]
			decoded := u.arena.DecodeString(seg.Span, u.charset)

			switch seg.Kind {
			case lang.SegEmpty:
				chain = append(chain, "[]")
			case lang.SegLiteral:
				chain = append(chain, "["+decoded+"]")
			default:
				if seg.Notation == lang.NotationRoot {
					chain = append(chain, decoded)
				} else {
					chain = append(chain, "["+decoded+"]")
				}
			}
		}

		// Extract value
		val := u.extractValue(param)

		// Wrap comma-split array if key ends with []
		if param.Key.SegLen > 0 {
			lastSeg := u.arena.Segments[int(param.Key.SegStart)+int(param.Key.SegLen)-1]
			if lastSeg.Kind == lang.SegEmpty {
				if arr, ok := val.([]any); ok {
					val = []any{arr}
				}
			}
		}

		// Build nested structure
		newObj := parseObject(chain, val, u.opts, true)
		if newObj != nil {
			merged := Merge(result, newObj)
			if m, ok := merged.(map[string]any); ok {
				result = m
			}
		}
	}

	// Compact if needed
	if !u.opts.AllowSparse {
		result = Compact(result).(map[string]any)
	}

	// Set result
	for k, v := range result {
		rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}

	return nil
}


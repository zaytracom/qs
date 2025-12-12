// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

// Benchmark comparison between popular Go query string libraries.
// Run: go test -bench=. -benchmem

package benchmarks

import (
	"net/url"
	"testing"

	ajgform "github.com/ajg/form"
	playgroundform "github.com/go-playground/form/v4"
	googleqs "github.com/google/go-querystring/query"
	"github.com/gorilla/schema"
	zaytraq "github.com/zaytracom/qs/v2"
)

// =============================================================================
// Test Data: Simple flat structure (all libraries can handle this)
// =============================================================================

type SimpleStruct struct {
	Name   string `form:"name" schema:"name" url:"name" qs:"name" query:"name"`
	Age    int    `form:"age" schema:"age" url:"age" qs:"age" query:"age"`
	Email  string `form:"email" schema:"email" url:"email" qs:"email" query:"email"`
	Active bool   `form:"active" schema:"active" url:"active" qs:"active" query:"active"`
}

var simpleStruct = SimpleStruct{
	Name:   "John",
	Age:    30,
	Email:  "john@example.com",
	Active: true,
}

var simpleValues = url.Values{
	"name":   []string{"John"},
	"age":    []string{"30"},
	"email":  []string{"john@example.com"},
	"active": []string{"true"},
}

var simpleQueryString = "name=John&age=30&email=john%40example.com&active=true"

// =============================================================================
// Test Data: Nested structure (only some libraries support)
// =============================================================================

type NestedProfile struct {
	Name string `form:"name" schema:"name" url:"name" qs:"name" query:"name"`
	Age  int    `form:"age" schema:"age" url:"age" qs:"age" query:"age"`
}

type NestedSettings struct {
	Theme string `form:"theme" schema:"theme" url:"theme" qs:"theme" query:"theme"`
	Lang  string `form:"lang" schema:"lang" url:"lang" qs:"lang" query:"lang"`
}

type NestedStruct struct {
	Profile  NestedProfile  `form:"profile" schema:"profile" url:"profile" qs:"profile" query:"profile"`
	Settings NestedSettings `form:"settings" schema:"settings" url:"settings" qs:"settings" query:"settings"`
}

var nestedStruct = NestedStruct{
	Profile:  NestedProfile{Name: "John", Age: 30},
	Settings: NestedSettings{Theme: "dark", Lang: "en"},
}

var nestedQueryString = "profile[name]=John&profile[age]=30&settings[theme]=dark&settings[lang]=en"

// =============================================================================
// Test Data: Array structure
// =============================================================================

type ArrayStruct struct {
	Tags []string `form:"tags" schema:"tags" url:"tags" qs:"tags" query:"tags"`
}

var arrayStruct = ArrayStruct{
	Tags: []string{"go", "rust", "python", "javascript", "typescript"},
}

// =============================================================================
// Test Data: Giant nested (for libraries that support it)
// =============================================================================

var giantNestedQueryString = "data[users][0][profile][name]=User0&data[users][0][profile][age]=20&data[users][0][settings][theme]=dark&data[users][0][settings][notifications][email]=true&data[users][0][settings][notifications][push]=false&" +
	"data[users][1][profile][name]=User1&data[users][1][profile][age]=21&data[users][1][settings][theme]=light&data[users][1][settings][notifications][email]=false&data[users][1][settings][notifications][push]=true&" +
	"data[users][2][profile][name]=User2&data[users][2][profile][age]=22&data[users][2][settings][theme]=auto&data[users][2][settings][notifications][email]=true&data[users][2][settings][notifications][push]=true&" +
	"data[config][api][version]=v2&data[config][api][timeout]=30&data[config][features][0]=feature1&data[config][features][1]=feature2"

var giantNestedData = map[string]any{
	"data": map[string]any{
		"users": []any{
			map[string]any{
				"profile":  map[string]any{"name": "User0", "age": 20},
				"settings": map[string]any{"theme": "dark", "notifications": map[string]any{"email": true, "push": false}},
			},
			map[string]any{
				"profile":  map[string]any{"name": "User1", "age": 21},
				"settings": map[string]any{"theme": "light", "notifications": map[string]any{"email": false, "push": true}},
			},
			map[string]any{
				"profile":  map[string]any{"name": "User2", "age": 22},
				"settings": map[string]any{"theme": "auto", "notifications": map[string]any{"email": true, "push": true}},
			},
		},
		"config": map[string]any{
			"api":      map[string]any{"version": "v2", "timeout": 30},
			"features": []any{"feature1", "feature2"},
		},
	},
}

// =============================================================================
// ENCODE Benchmarks: Simple struct
// =============================================================================

func BenchmarkEncode_Simple_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := zaytraq.Stringify(simpleStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Simple_GorillaSchema(b *testing.B) {
	encoder := schema.NewEncoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values := url.Values{}
		err := encoder.Encode(simpleStruct, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Simple_PlaygroundForm(b *testing.B) {
	encoder := playgroundform.NewEncoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encoder.Encode(&simpleStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Simple_AjgForm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ajgform.EncodeToValues(simpleStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Simple_GoogleQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := googleqs.Values(simpleStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}


// =============================================================================
// ENCODE Benchmarks: Nested struct (libraries that support it)
// =============================================================================

func BenchmarkEncode_Nested_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := zaytraq.Stringify(nestedStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Nested_GorillaSchema(b *testing.B) {
	b.Skip("gorilla/schema does not support nested objects encoding")
}

func BenchmarkEncode_Nested_PlaygroundForm(b *testing.B) {
	encoder := playgroundform.NewEncoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encoder.Encode(&nestedStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Nested_AjgForm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ajgform.EncodeToValues(nestedStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Nested_GoogleQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := googleqs.Values(nestedStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}


// =============================================================================
// ENCODE Benchmarks: Array
// =============================================================================

func BenchmarkEncode_Array_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := zaytraq.Stringify(arrayStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Array_GorillaSchema(b *testing.B) {
	b.Skip("gorilla/schema does not support array encoding with brackets")
}

func BenchmarkEncode_Array_PlaygroundForm(b *testing.B) {
	encoder := playgroundform.NewEncoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encoder.Encode(&arrayStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Array_AjgForm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ajgform.EncodeToValues(arrayStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Array_GoogleQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := googleqs.Values(arrayStruct)
		if err != nil {
			b.Fatal(err)
		}
	}
}


// =============================================================================
// ENCODE Benchmarks: Giant nested map
// =============================================================================

func BenchmarkEncode_Giant_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := zaytraq.Stringify(giantNestedData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode_Giant_GorillaSchema(b *testing.B) {
	b.Skip("gorilla/schema does not support dynamic map[string]any encoding")
}

func BenchmarkEncode_Giant_PlaygroundForm(b *testing.B) {
	b.Skip("go-playground/form does not support dynamic map[string]any encoding")
}

func BenchmarkEncode_Giant_AjgForm(b *testing.B) {
	b.Skip("ajg/form does not support dynamic map[string]any encoding")
}

func BenchmarkEncode_Giant_GoogleQS(b *testing.B) {
	b.Skip("google/go-querystring does not support dynamic map[string]any encoding")
}

// =============================================================================
// DECODE Benchmarks: Simple struct
// =============================================================================

func BenchmarkDecode_Simple_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s SimpleStruct
		err := zaytraq.Unmarshal(simpleQueryString, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Simple_GorillaSchema(b *testing.B) {
	decoder := schema.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SimpleStruct
		err := decoder.Decode(&s, simpleValues)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Simple_PlaygroundForm(b *testing.B) {
	decoder := playgroundform.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SimpleStruct
		err := decoder.Decode(&s, simpleValues)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Simple_AjgForm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s SimpleStruct
		err := ajgform.DecodeValues(&s, simpleValues)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Simple_GoogleQS(b *testing.B) {
	b.Skip("google/go-querystring does not support decoding")
}

// =============================================================================
// DECODE Benchmarks: Nested struct
// =============================================================================

var nestedValues = url.Values{
	"profile[name]":   []string{"John"},
	"profile[age]":    []string{"30"},
	"settings[theme]": []string{"dark"},
	"settings[lang]":  []string{"en"},
}

func BenchmarkDecode_Nested_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s NestedStruct
		err := zaytraq.Unmarshal(nestedQueryString, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Nested_GorillaSchema(b *testing.B) {
	b.Skip("gorilla/schema does not support nested decoding")
}

func BenchmarkDecode_Nested_PlaygroundForm(b *testing.B) {
	decoder := playgroundform.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s NestedStruct
		err := decoder.Decode(&s, nestedValues)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Nested_AjgForm(b *testing.B) {
	b.Skip("ajg/form uses dot notation for nested, not brackets")
}

func BenchmarkDecode_Nested_GoogleQS(b *testing.B) {
	b.Skip("google/go-querystring does not support decoding")
}

// =============================================================================
// DECODE Benchmarks: Array struct
// =============================================================================

var arrayValues = url.Values{
	"tags": []string{"go", "rust", "python", "javascript", "typescript"},
}

var arrayQueryString = "tags[0]=go&tags[1]=rust&tags[2]=python&tags[3]=javascript&tags[4]=typescript"

func BenchmarkDecode_Array_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s ArrayStruct
		err := zaytraq.Unmarshal(arrayQueryString, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Array_GorillaSchema(b *testing.B) {
	b.Skip("gorilla/schema does not support array decoding with brackets")
}

func BenchmarkDecode_Array_PlaygroundForm(b *testing.B) {
	decoder := playgroundform.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s ArrayStruct
		err := decoder.Decode(&s, arrayValues)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_Array_AjgForm(b *testing.B) {
	b.Skip("ajg/form uses indices notation for arrays")
}

func BenchmarkDecode_Array_GoogleQS(b *testing.B) {
	b.Skip("google/go-querystring does not support decoding")
}

// =============================================================================
// DECODE Benchmarks: Dynamic map (only zaytra supports)
// =============================================================================

func BenchmarkDecode_DynamicMap_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := zaytraq.Parse(giantNestedQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode_DynamicMap_GorillaSchema(b *testing.B) {
	b.Skip("gorilla/schema does not support dynamic map parsing")
}

func BenchmarkDecode_DynamicMap_PlaygroundForm(b *testing.B) {
	b.Skip("go-playground/form does not support dynamic map parsing")
}

func BenchmarkDecode_DynamicMap_AjgForm(b *testing.B) {
	b.Skip("ajg/form does not support dynamic map parsing")
}

func BenchmarkDecode_DynamicMap_GoogleQS(b *testing.B) {
	b.Skip("google/go-querystring does not support decoding")
}

// =============================================================================
// FAIR DECODE Benchmarks: All libraries parse from raw query string
// Each library uses its native format for fair comparison
// =============================================================================

// Query strings in each library's native format
var (
	// Simple: all libraries use the same flat format
	// "name=John&age=30&email=john%40example.com&active=true"

	// Nested: different formats per library
	nestedQueryStringDot      = "profile.name=John&profile.age=30&settings.theme=dark&settings.lang=en"       // go-playground, ajg
	nestedQueryStringBracket  = "profile[name]=John&profile[age]=30&settings[theme]=dark&settings[lang]=en"   // zaytra

	// Array: different formats per library
	arrayQueryStringRepeat  = "tags=go&tags=rust&tags=python&tags=javascript&tags=typescript"                           // gorilla, go-playground
	arrayQueryStringIndices = "tags[0]=go&tags[1]=rust&tags[2]=python&tags[3]=javascript&tags[4]=typescript"            // zaytra
	arrayQueryStringDot     = "tags.0=go&tags.1=rust&tags.2=python&tags.3=javascript&tags.4=typescript"                 // ajg
)

// =============================================================================
// FAIR DECODE: Simple struct (all use same format)
// =============================================================================

func BenchmarkFairDecode_Simple_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s SimpleStruct
		err := zaytraq.Unmarshal(simpleQueryString, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Simple_GorillaSchema(b *testing.B) {
	decoder := schema.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(simpleQueryString)
		if err != nil {
			b.Fatal(err)
		}
		var s SimpleStruct
		err = decoder.Decode(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Simple_PlaygroundForm(b *testing.B) {
	decoder := playgroundform.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(simpleQueryString)
		if err != nil {
			b.Fatal(err)
		}
		var s SimpleStruct
		err = decoder.Decode(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Simple_AjgForm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(simpleQueryString)
		if err != nil {
			b.Fatal(err)
		}
		var s SimpleStruct
		err = ajgform.DecodeValues(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// FAIR DECODE: Nested struct (each library uses its native format)
// =============================================================================

func BenchmarkFairDecode_Nested_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s NestedStruct
		err := zaytraq.Unmarshal(nestedQueryStringBracket, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Nested_GorillaSchema(b *testing.B) {
	b.Skip("gorilla/schema does not support nested struct decoding")
}

func BenchmarkFairDecode_Nested_PlaygroundForm(b *testing.B) {
	decoder := playgroundform.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(nestedQueryStringDot)
		if err != nil {
			b.Fatal(err)
		}
		var s NestedStruct
		err = decoder.Decode(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Nested_AjgForm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(nestedQueryStringDot)
		if err != nil {
			b.Fatal(err)
		}
		var s NestedStruct
		err = ajgform.DecodeValues(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// FAIR DECODE: Array struct (each library uses its native format)
// =============================================================================

func BenchmarkFairDecode_Array_ZaytraQS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s ArrayStruct
		err := zaytraq.Unmarshal(arrayQueryStringIndices, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Array_GorillaSchema(b *testing.B) {
	decoder := schema.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(arrayQueryStringRepeat)
		if err != nil {
			b.Fatal(err)
		}
		var s ArrayStruct
		err = decoder.Decode(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Array_PlaygroundForm(b *testing.B) {
	decoder := playgroundform.NewDecoder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(arrayQueryStringRepeat)
		if err != nil {
			b.Fatal(err)
		}
		var s ArrayStruct
		err = decoder.Decode(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFairDecode_Array_AjgForm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		values, err := url.ParseQuery(arrayQueryStringDot)
		if err != nil {
			b.Fatal(err)
		}
		var s ArrayStruct
		err = ajgform.DecodeValues(&s, values)
		if err != nil {
			b.Fatal(err)
		}
	}
}

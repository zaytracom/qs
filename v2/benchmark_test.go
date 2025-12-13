// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"testing"
)

// =============================================================================
// Test Data: Simple flat structure
// =============================================================================

var simpleQueryString = "name=John&age=30&email=john@example.com&active=true"

var simpleData = map[string]any{
	"name":   "John",
	"age":    "30",
	"email":  "john@example.com",
	"active": "true",
}

// =============================================================================
// Test Data: Nested structure (3 levels deep)
// =============================================================================

var nestedQueryString = "user[profile][name]=John&user[profile][age]=30&user[settings][theme]=dark&user[settings][lang]=en"

var nestedData = map[string]any{
	"user": map[string]any{
		"profile": map[string]any{
			"name": "John",
			"age":  30,
		},
		"settings": map[string]any{
			"theme": "dark",
			"lang":  "en",
		},
	},
}

// =============================================================================
// Test Data: Array with indices
// =============================================================================

var arrayQueryString = "items[0]=apple&items[1]=banana&items[2]=cherry&items[3]=date&items[4]=elderberry"

var arrayData = map[string]any{
	"items": []any{"apple", "banana", "cherry", "date", "elderberry"},
}

// =============================================================================
// Test Data: Deep nested (5 levels)
// =============================================================================

var deepNestedQueryString = "a[b][c][d][e]=value&a[b][c][d][f]=other&a[b][c][g]=test&a[b][h]=data&a[i]=root"

var deepNestedData = map[string]any{
	"a": map[string]any{
		"b": map[string]any{
			"c": map[string]any{
				"d": map[string]any{
					"e": "value",
					"f": "other",
				},
				"g": "test",
			},
			"h": "data",
		},
		"i": "root",
	},
}

// =============================================================================
// Test Data: Complex - nested with arrays (Strapi-like)
// =============================================================================

var complexQueryString = "filters[status][$eq]=published&filters[author][name][$contains]=john&sort[0]=createdAt:desc&sort[1]=title:asc&populate[author][fields][0]=name&populate[author][fields][1]=email&pagination[page]=1&pagination[pageSize]=25"

var complexData = map[string]any{
	"filters": map[string]any{
		"status": map[string]any{
			"$eq": "published",
		},
		"author": map[string]any{
			"name": map[string]any{
				"$contains": "john",
			},
		},
	},
	"sort": []any{"createdAt:desc", "title:asc"},
	"populate": map[string]any{
		"author": map[string]any{
			"fields": []any{"name", "email"},
		},
	},
	"pagination": map[string]any{
		"page":     1,
		"pageSize": 25,
	},
}

// =============================================================================
// Test Data: Giant nested structure (stress test)
// =============================================================================

func generateGiantNestedQueryString() string {
	// 10 top-level keys, each with 5 nested levels and arrays
	return "data[users][0][profile][name]=User0&data[users][0][profile][age]=20&data[users][0][settings][theme]=dark&data[users][0][settings][notifications][email]=true&data[users][0][settings][notifications][push]=false&" +
		"data[users][1][profile][name]=User1&data[users][1][profile][age]=21&data[users][1][settings][theme]=light&data[users][1][settings][notifications][email]=false&data[users][1][settings][notifications][push]=true&" +
		"data[users][2][profile][name]=User2&data[users][2][profile][age]=22&data[users][2][settings][theme]=auto&data[users][2][settings][notifications][email]=true&data[users][2][settings][notifications][push]=true&" +
		"data[users][3][profile][name]=User3&data[users][3][profile][age]=23&data[users][3][settings][theme]=dark&data[users][3][settings][notifications][email]=false&data[users][3][settings][notifications][push]=false&" +
		"data[users][4][profile][name]=User4&data[users][4][profile][age]=24&data[users][4][settings][theme]=light&data[users][4][settings][notifications][email]=true&data[users][4][settings][notifications][push]=false&" +
		"data[products][0][info][title]=Product0&data[products][0][info][price]=100&data[products][0][meta][tags][0]=tag1&data[products][0][meta][tags][1]=tag2&data[products][0][meta][category][name]=Cat0&" +
		"data[products][1][info][title]=Product1&data[products][1][info][price]=200&data[products][1][meta][tags][0]=tag3&data[products][1][meta][tags][1]=tag4&data[products][1][meta][category][name]=Cat1&" +
		"data[products][2][info][title]=Product2&data[products][2][info][price]=300&data[products][2][meta][tags][0]=tag5&data[products][2][meta][tags][1]=tag6&data[products][2][meta][category][name]=Cat2&" +
		"data[config][api][version]=v2&data[config][api][timeout]=30&data[config][api][retries]=3&data[config][features][0]=feature1&data[config][features][1]=feature2&data[config][features][2]=feature3"
}

func generateGiantNestedData() map[string]any {
	users := make([]any, 5)
	for i := 0; i < 5; i++ {
		users[i] = map[string]any{
			"profile": map[string]any{
				"name": "User" + string(rune('0'+i)),
				"age":  20 + i,
			},
			"settings": map[string]any{
				"theme": []string{"dark", "light", "auto", "dark", "light"}[i],
				"notifications": map[string]any{
					"email": i%2 == 0,
					"push":  i%2 == 1,
				},
			},
		}
	}

	products := make([]any, 3)
	for i := 0; i < 3; i++ {
		products[i] = map[string]any{
			"info": map[string]any{
				"title": "Product" + string(rune('0'+i)),
				"price": (i + 1) * 100,
			},
			"meta": map[string]any{
				"tags": []any{"tag" + string(rune('1'+i*2)), "tag" + string(rune('2'+i*2))},
				"category": map[string]any{
					"name": "Cat" + string(rune('0'+i)),
				},
			},
		}
	}

	return map[string]any{
		"data": map[string]any{
			"users":    users,
			"products": products,
			"config": map[string]any{
				"api": map[string]any{
					"version": "v2",
					"timeout": 30,
					"retries": 3,
				},
				"features": []any{"feature1", "feature2", "feature3"},
			},
		},
	}
}

var giantNestedQueryString = generateGiantNestedQueryString()
var giantNestedData = generateGiantNestedData()

// =============================================================================
// Benchmarks: Parse
// =============================================================================

func BenchmarkParse_Simple(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Parse(simpleQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_Nested(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Parse(nestedQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_Array(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Parse(arrayQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_DeepNested(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Parse(deepNestedQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_Complex(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Parse(complexQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_Giant(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Parse(giantNestedQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// Benchmarks: Stringify
// =============================================================================

func BenchmarkStringify_Simple(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Stringify(simpleData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringify_Nested(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Stringify(nestedData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringify_Array(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Stringify(arrayData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringify_DeepNested(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Stringify(deepNestedData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringify_Complex(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Stringify(complexData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringify_Giant(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Stringify(giantNestedData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// Benchmarks: Parallel
// =============================================================================

func BenchmarkParse_Simple_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Parse(simpleQueryString)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkParse_Giant_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Parse(giantNestedQueryString)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkStringify_Simple_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Stringify(simpleData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkStringify_Giant_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Stringify(giantNestedData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// =============================================================================
// Benchmarks: UnmarshalBytes (direct AST → struct)
// =============================================================================

type SimpleStruct struct {
	Name   string `query:"name"`
	Age    string `query:"age"`
	Email  string `query:"email"`
	Active string `query:"active"`
}

type NestedStruct struct {
	User struct {
		Profile struct {
			Name string `query:"name"`
			Age  string `query:"age"`
		} `query:"profile"`
		Settings struct {
			Theme string `query:"theme"`
			Lang  string `query:"lang"`
		} `query:"settings"`
	} `query:"user"`
}

type ArrayStruct struct {
	Items []string `query:"items"`
}

func BenchmarkUnmarshalBytes_Simple_Struct(b *testing.B) {
	data := []byte(simpleQueryString)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s SimpleStruct
		err := UnmarshalBytes(data, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalBytes_Simple_Map(b *testing.B) {
	data := []byte(simpleQueryString)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var m map[string]any
		err := UnmarshalBytes(data, &m)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalBytes_Nested_Struct(b *testing.B) {
	data := []byte(nestedQueryString)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s NestedStruct
		err := UnmarshalBytes(data, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalBytes_Array_Struct(b *testing.B) {
	data := []byte(arrayQueryString)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s ArrayStruct
		err := UnmarshalBytes(data, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalBytes_Giant_Map(b *testing.B) {
	data := []byte(giantNestedQueryString)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var m map[string]any
		err := UnmarshalBytes(data, &m)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Comparison: Parse (old way) vs UnmarshalBytes (new way)
func BenchmarkCompare_Simple_Parse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Parse(simpleQueryString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompare_Simple_UnmarshalMap(b *testing.B) {
	data := []byte(simpleQueryString)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var m map[string]any
		err := UnmarshalBytes(data, &m)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompare_Simple_UnmarshalStruct(b *testing.B) {
	data := []byte(simpleQueryString)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s SimpleStruct
		err := UnmarshalBytes(data, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

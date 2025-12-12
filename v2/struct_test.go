// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"reflect"
	"testing"
	"time"
)

// Test struct types
type SimpleUser struct {
	Name  string `query:"name"`
	Age   int    `query:"age"`
	Email string `query:"email"`
}

type UserWithDefaults struct {
	Name     string
	Age      int
	IsActive bool
}

type UserWithAllTypes struct {
	Name     string  `query:"name"`
	Age      int     `query:"age"`
	Score    float64 `query:"score"`
	Active   bool    `query:"active"`
	Count    int64   `query:"count"`
	ID       uint    `query:"id"`
	Rating   float32 `query:"rating"`
}

type NestedAddress struct {
	Street  string `query:"street"`
	City    string `query:"city"`
	ZipCode string `query:"zip"`
}

type UserWithNested struct {
	Name    string        `query:"name"`
	Address NestedAddress `query:"address"`
}

type UserWithSlice struct {
	Name   string   `query:"name"`
	Tags   []string `query:"tags"`
	Scores []int    `query:"scores"`
}

type UserWithMap struct {
	Name    string            `query:"name"`
	Profile map[string]string `query:"profile"`
}

type UserWithPointer struct {
	Name  string  `query:"name"`
	Age   *int    `query:"age"`
	Email *string `query:"email"`
}

type UserWithTime struct {
	Name      string    `query:"name"`
	CreatedAt time.Time `query:"created_at"`
}

type UserWithSkip struct {
	Name     string `query:"name"`
	Password string `query:"-"`
	Email    string `query:"email"`
}

type UserWithAny struct {
	Name    string `query:"name"`
	Payload any    `query:"payload"`
}

// TestParseToStruct tests basic struct parsing
func TestParseToStruct(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		dest     any
		expected any
		wantErr  bool
	}{
		{
			name:  "simple struct with query tags",
			input: "name=John&age=30&email=john@example.com",
			dest:  &SimpleUser{},
			expected: &SimpleUser{
				Name:  "John",
				Age:   30,
				Email: "john@example.com",
			},
		},
		{
			name:  "struct with default field names (lowercase)",
			input: "name=Jane&age=25&isactive=true",
			dest:  &UserWithDefaults{},
			expected: &UserWithDefaults{
				Name:     "Jane",
				Age:      25,
				IsActive: true,
			},
		},
		{
			name:  "struct with all types",
			input: "name=Test&age=42&score=95.5&active=true&count=1000&id=123&rating=4.5",
			dest:  &UserWithAllTypes{},
			expected: &UserWithAllTypes{
				Name:   "Test",
				Age:    42,
				Score:  95.5,
				Active: true,
				Count:  1000,
				ID:     123,
				Rating: 4.5,
			},
		},
		{
			name:  "struct with nested object",
			input: "name=John&address[street]=123+Main+St&address[city]=NYC&address[zip]=10001",
			dest:  &UserWithNested{},
			expected: &UserWithNested{
				Name: "John",
				Address: NestedAddress{
					Street:  "123 Main St",
					City:    "NYC",
					ZipCode: "10001",
				},
			},
		},
		{
			name:  "struct with slice",
			input: "name=John&tags[]=go&tags[]=programming&scores[]=100&scores[]=90",
			dest:  &UserWithSlice{},
			expected: &UserWithSlice{
				Name:   "John",
				Tags:   []string{"go", "programming"},
				Scores: []int{100, 90},
			},
		},
		{
			name:  "struct with map",
			input: "name=John&profile[bio]=developer&profile[location]=NYC",
			dest:  &UserWithMap{},
			expected: &UserWithMap{
				Name: "John",
				Profile: map[string]string{
					"bio":      "developer",
					"location": "NYC",
				},
			},
		},
		{
			name:  "struct with skip tag",
			input: "name=John&password=secret&email=john@example.com",
			dest:  &UserWithSkip{},
			expected: &UserWithSkip{
				Name:     "John",
				Password: "", // Should be skipped
				Email:    "john@example.com",
			},
		},
		{
			name:     "empty input",
			input:    "",
			dest:     &SimpleUser{},
			expected: &SimpleUser{},
		},
		{
			name:  "partial fields",
			input: "name=John",
			dest:  &SimpleUser{},
			expected: &SimpleUser{
				Name: "John",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseToStruct(tt.input, tt.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.dest, tt.expected) {
				t.Errorf("ParseToStruct() = %+v, want %+v", tt.dest, tt.expected)
			}
		})
	}
}

// TestParseToStructPointers tests pointer field handling
func TestParseToStructPointers(t *testing.T) {
	t.Run("pointer fields are set", func(t *testing.T) {
		var user UserWithPointer
		err := ParseToStruct("name=John&age=30&email=john@example.com", &user)
		if err != nil {
			t.Fatalf("ParseToStruct() error = %v", err)
		}

		if user.Name != "John" {
			t.Errorf("Name = %v, want John", user.Name)
		}
		if user.Age == nil || *user.Age != 30 {
			t.Errorf("Age = %v, want 30", user.Age)
		}
		if user.Email == nil || *user.Email != "john@example.com" {
			t.Errorf("Email = %v, want john@example.com", user.Email)
		}
	})

	t.Run("nil pointer fields when not provided", func(t *testing.T) {
		var user UserWithPointer
		err := ParseToStruct("name=John", &user)
		if err != nil {
			t.Fatalf("ParseToStruct() error = %v", err)
		}

		if user.Name != "John" {
			t.Errorf("Name = %v, want John", user.Name)
		}
		if user.Age != nil {
			t.Errorf("Age = %v, want nil", user.Age)
		}
		if user.Email != nil {
			t.Errorf("Email = %v, want nil", user.Email)
		}
	})
}

// TestParseToStructTime tests time.Time field handling
func TestParseToStructTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "RFC3339 format",
			input:   "name=John&created_at=2024-01-15T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "RFC3339 with timezone",
			input:   "name=John&created_at=2024-01-15T10:30:00-05:00",
			wantErr: false,
		},
		{
			name:    "date only format",
			input:   "name=John&created_at=2024-01-15",
			wantErr: false,
		},
		{
			name:    "invalid time format",
			input:   "name=John&created_at=invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user UserWithTime
			err := ParseToStruct(tt.input, &user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && user.Name != "John" {
				t.Errorf("Name = %v, want John", user.Name)
			}
		})
	}
}

// TestParseToStructErrors tests error cases
func TestParseToStructErrors(t *testing.T) {
	tests := []struct {
		name    string
		dest    any
		wantErr bool
	}{
		{
			name:    "nil destination",
			dest:    nil,
			wantErr: true,
		},
		{
			name:    "non-pointer destination",
			dest:    SimpleUser{},
			wantErr: true,
		},
		{
			name:    "pointer to non-struct",
			dest:    new(string),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseToStruct("name=John", tt.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStructToQueryString tests struct stringification
func TestStructToQueryString(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name: "simple struct",
			input: SimpleUser{
				Name:  "John",
				Age:   30,
				Email: "john@example.com",
			},
			// Note: order may vary, so we'll check key presence
			wantErr: false,
		},
		{
			name: "struct with zero values",
			input: SimpleUser{
				Name: "John",
			},
			wantErr: false,
		},
		{
			name: "struct with skip tag",
			input: UserWithSkip{
				Name:     "John",
				Password: "secret",
				Email:    "john@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StructToQueryString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructToQueryString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" && tt.input != nil {
				// Check that we got some output for non-nil inputs
				if s, ok := tt.input.(SimpleUser); ok && s.Name != "" {
					// Should have at least the name
					if result == "" {
						t.Errorf("StructToQueryString() returned empty for non-empty struct")
					}
				}
			}
		})
	}
}

// TestStructToQueryStringSkipTag verifies skip tag works
func TestStructToQueryStringSkipTag(t *testing.T) {
	user := UserWithSkip{
		Name:     "John",
		Password: "secret",
		Email:    "john@example.com",
	}

	result, err := StructToQueryString(user)
	if err != nil {
		t.Fatalf("StructToQueryString() error = %v", err)
	}

	// Parse back to verify password is not included
	parsed, _ := Parse(result)

	if _, ok := parsed["password"]; ok {
		t.Error("password should not be in query string (has skip tag)")
	}
	if name, ok := parsed["name"].(string); !ok || name != "John" {
		t.Errorf("name = %v, want John", parsed["name"])
	}
}

// TestMarshalUnmarshal tests the Marshal and Unmarshal functions
func TestMarshalUnmarshal(t *testing.T) {
	t.Run("struct roundtrip", func(t *testing.T) {
		original := SimpleUser{
			Name:  "John",
			Age:   30,
			Email: "john@example.com",
		}

		// Marshal
		str, err := Marshal(original)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		// Unmarshal
		var result SimpleUser
		err = Unmarshal(str, &result)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if result.Name != original.Name || result.Age != original.Age || result.Email != original.Email {
			t.Errorf("Unmarshal() = %+v, want %+v", result, original)
		}
	})

	t.Run("map roundtrip", func(t *testing.T) {
		original := map[string]any{
			"name": "John",
			"age":  30,
		}

		// Marshal
		str, err := Marshal(original)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		// Unmarshal
		var result map[string]any
		err = Unmarshal(str, &result)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if result["name"] != "John" {
			t.Errorf("result[name] = %v, want John", result["name"])
		}
	})

	t.Run("unmarshal to interface", func(t *testing.T) {
		var result any
		err := Unmarshal("name=John&age=30", &result)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("result is not map[string]any")
		}
		if m["name"] != "John" {
			t.Errorf("result[name] = %v, want John", m["name"])
		}
	})

	t.Run("marshal nil", func(t *testing.T) {
		str, err := Marshal(nil)
		if err != nil {
			t.Fatalf("Marshal(nil) error = %v", err)
		}
		if str != "" {
			t.Errorf("Marshal(nil) = %q, want empty string", str)
		}
	})

	t.Run("unmarshal errors", func(t *testing.T) {
		// nil target
		err := Unmarshal("name=John", nil)
		if err == nil {
			t.Error("Unmarshal(nil target) should return error")
		}

		// non-pointer target
		var user SimpleUser
		err = Unmarshal("name=John", user)
		if err == nil {
			t.Error("Unmarshal(non-pointer) should return error")
		}
	})
}

// TestMarshalNestedStruct tests marshaling nested structs
func TestMarshalNestedStruct(t *testing.T) {
	user := UserWithNested{
		Name: "John",
		Address: NestedAddress{
			Street:  "123 Main St",
			City:    "NYC",
			ZipCode: "10001",
		},
	}

	str, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal back
	var result UserWithNested
	err = Unmarshal(str, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.Name != user.Name {
		t.Errorf("Name = %v, want %v", result.Name, user.Name)
	}
	if result.Address.Street != user.Address.Street {
		t.Errorf("Address.Street = %v, want %v", result.Address.Street, user.Address.Street)
	}
	if result.Address.City != user.Address.City {
		t.Errorf("Address.City = %v, want %v", result.Address.City, user.Address.City)
	}
}

// TestMarshalSliceStruct tests marshaling structs with slices
func TestMarshalSliceStruct(t *testing.T) {
	user := UserWithSlice{
		Name:   "John",
		Tags:   []string{"go", "programming"},
		Scores: []int{100, 90, 85},
	}

	str, err := Marshal(user)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal back
	var result UserWithSlice
	err = Unmarshal(str, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.Name != user.Name {
		t.Errorf("Name = %v, want %v", result.Name, user.Name)
	}
	if !reflect.DeepEqual(result.Tags, user.Tags) {
		t.Errorf("Tags = %v, want %v", result.Tags, user.Tags)
	}
	if !reflect.DeepEqual(result.Scores, user.Scores) {
		t.Errorf("Scores = %v, want %v", result.Scores, user.Scores)
	}
}

// TestStructToMap tests StructToMap conversion
func TestStructToMap(t *testing.T) {
	user := SimpleUser{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}

	m, err := StructToMap(user)
	if err != nil {
		t.Fatalf("StructToMap() error = %v", err)
	}

	if m["name"] != "John" {
		t.Errorf("m[name] = %v, want John", m["name"])
	}
	if m["age"] != 30 {
		t.Errorf("m[age] = %v, want 30", m["age"])
	}
	if m["email"] != "john@example.com" {
		t.Errorf("m[email] = %v, want john@example.com", m["email"])
	}
}

// TestMapToStruct tests MapToStruct conversion
func TestMapToStruct(t *testing.T) {
	data := map[string]any{
		"name":  "John",
		"age":   "30",
		"email": "john@example.com",
	}

	var user SimpleUser
	err := MapToStruct(data, &user)
	if err != nil {
		t.Fatalf("MapToStruct() error = %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Name = %v, want John", user.Name)
	}
	if user.Age != 30 {
		t.Errorf("Age = %v, want 30", user.Age)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Email = %v, want john@example.com", user.Email)
	}
}

// TestParseToStructWithOptions tests ParseToStruct with parse options
func TestParseToStructWithOptions(t *testing.T) {
	type DotNotationUser struct {
		Name    string `query:"name"`
		Profile struct {
			Bio string `query:"bio"`
		} `query:"profile"`
	}

	var user DotNotationUser
	err := ParseToStruct("name=John&profile.bio=Developer", &user, WithAllowDots(true))
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Name = %v, want John", user.Name)
	}
	if user.Profile.Bio != "Developer" {
		t.Errorf("Profile.Bio = %v, want Developer", user.Profile.Bio)
	}
}

// TestMarshalWithOptions tests Marshal with stringify options
func TestMarshalWithOptions(t *testing.T) {
	user := UserWithSlice{
		Name: "John",
		Tags: []string{"go", "programming"},
	}

	// Test with brackets format
	str, err := Marshal(user, WithArrayFormat(ArrayFormatBrackets))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Should contain brackets format
	parsed, _ := Parse(str)
	if parsed["name"] != "John" {
		t.Errorf("Parsed name = %v, want John", parsed["name"])
	}
}

// TestTypeConversions tests various type conversions
func TestTypeConversions(t *testing.T) {
	type AllTypes struct {
		String  string  `query:"string"`
		Int     int     `query:"int"`
		Int8    int8    `query:"int8"`
		Int16   int16   `query:"int16"`
		Int32   int32   `query:"int32"`
		Int64   int64   `query:"int64"`
		Uint    uint    `query:"uint"`
		Uint8   uint8   `query:"uint8"`
		Uint16  uint16  `query:"uint16"`
		Uint32  uint32  `query:"uint32"`
		Uint64  uint64  `query:"uint64"`
		Float32 float32 `query:"float32"`
		Float64 float64 `query:"float64"`
		Bool    bool    `query:"bool"`
	}

	input := "string=hello&int=42&int8=8&int16=16&int32=32&int64=64" +
		"&uint=1&uint8=2&uint16=3&uint32=4&uint64=5" +
		"&float32=1.5&float64=2.5&bool=true"

	var result AllTypes
	err := ParseToStruct(input, &result)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	if result.String != "hello" {
		t.Errorf("String = %v, want hello", result.String)
	}
	if result.Int != 42 {
		t.Errorf("Int = %v, want 42", result.Int)
	}
	if result.Int8 != 8 {
		t.Errorf("Int8 = %v, want 8", result.Int8)
	}
	if result.Float32 != 1.5 {
		t.Errorf("Float32 = %v, want 1.5", result.Float32)
	}
	if result.Float64 != 2.5 {
		t.Errorf("Float64 = %v, want 2.5", result.Float64)
	}
	if result.Bool != true {
		t.Errorf("Bool = %v, want true", result.Bool)
	}
}

// TestEmptyStringConversions tests that empty strings don't cause errors
func TestEmptyStringConversions(t *testing.T) {
	type Numbers struct {
		Int   int     `query:"int"`
		Float float64 `query:"float"`
		Bool  bool    `query:"bool"`
	}

	// Empty values should not cause conversion errors
	var result Numbers
	err := ParseToStruct("int=&float=&bool=", &result)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	// All fields should remain zero values
	if result.Int != 0 {
		t.Errorf("Int = %v, want 0", result.Int)
	}
	if result.Float != 0 {
		t.Errorf("Float = %v, want 0", result.Float)
	}
	if result.Bool != false {
		t.Errorf("Bool = %v, want false", result.Bool)
	}
}

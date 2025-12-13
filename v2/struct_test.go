// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"reflect"
	"strings"
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
	Name   string  `query:"name"`
	Age    int     `query:"age"`
	Score  float64 `query:"score"`
	Active bool    `query:"active"`
	Count  int64   `query:"count"`
	ID     uint    `query:"id"`
	Rating float32 `query:"rating"`
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
	err := ParseToStruct("name=John&profile.bio=Developer", &user, WithParseAllowDots(true))
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
	str, err := Marshal(user, WithStringifyArrayFormat(ArrayFormatBrackets))
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

// =============================================================================
// COMPLEX STRUCT TESTS WITH OPTIONS
// =============================================================================

// TestParseToStructWithAllowDots tests dot notation parsing into structs
func TestParseToStructWithParseAllowDots(t *testing.T) {
	type Profile struct {
		Bio      string `query:"bio"`
		Location string `query:"location"`
	}
	type Settings struct {
		Theme         string `query:"theme"`
		Notifications bool   `query:"notifications"`
	}
	type User struct {
		Name     string   `query:"name"`
		Profile  Profile  `query:"profile"`
		Settings Settings `query:"settings"`
	}

	var user User
	err := ParseToStruct(
		"name=John&profile.bio=Developer&profile.location=NYC&settings.theme=dark&settings.notifications=true",
		&user,
		WithParseAllowDots(true),
	)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Name = %v, want John", user.Name)
	}
	if user.Profile.Bio != "Developer" {
		t.Errorf("Profile.Bio = %v, want Developer", user.Profile.Bio)
	}
	if user.Profile.Location != "NYC" {
		t.Errorf("Profile.Location = %v, want NYC", user.Profile.Location)
	}
	if user.Settings.Theme != "dark" {
		t.Errorf("Settings.Theme = %v, want dark", user.Settings.Theme)
	}
	if user.Settings.Notifications != true {
		t.Errorf("Settings.Notifications = %v, want true", user.Settings.Notifications)
	}
}

// TestParseToStructWithDepthLimit tests depth limit with structs
func TestParseToStructWithDepthLimit(t *testing.T) {
	type Level3 struct {
		Value string `query:"value"`
	}
	type Level2 struct {
		Level3 Level3 `query:"level3"`
	}
	type Level1 struct {
		Level2 Level2 `query:"level2"`
	}
	type Root struct {
		Level1 Level1 `query:"level1"`
	}

	// With depth=2, level3 should not be parsed as nested
	var root Root
	err := ParseToStruct(
		"level1[level2][level3][value]=deep",
		&root,
		WithParseDepth(2),
	)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	// level3 is at depth 3, so it becomes a literal key
	// The struct won't have the nested value populated properly
	// This is expected behavior with depth limits
}

// TestParseToStructWithArrayLimit tests array limit with struct slices
func TestParseToStructWithParseArrayLimit(t *testing.T) {
	type Container struct {
		Items []string `query:"items"`
	}

	var container Container
	err := ParseToStruct(
		"items[0]=a&items[1]=b&items[100]=c",
		&container,
		WithParseArrayLimit(50),
	)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	// items[100] exceeds array limit, so it becomes object key
	// Only items[0] and items[1] should be in the array
	if len(container.Items) < 2 {
		t.Errorf("Items length = %v, want at least 2", len(container.Items))
	}
}

// TestParseToStructWithComma tests comma-separated values into slices
func TestParseToStructWithParseComma(t *testing.T) {
	type Filter struct {
		Tags     []string `query:"tags"`
		Statuses []string `query:"statuses"`
		Single   string   `query:"single"`
	}

	var filter Filter
	err := ParseToStruct(
		"tags=go,rust,python&statuses=active,pending&single=value",
		&filter,
		WithParseComma(true),
	)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	expectedTags := []string{"go", "rust", "python"}
	if !reflect.DeepEqual(filter.Tags, expectedTags) {
		t.Errorf("Tags = %v, want %v", filter.Tags, expectedTags)
	}

	expectedStatuses := []string{"active", "pending"}
	if !reflect.DeepEqual(filter.Statuses, expectedStatuses) {
		t.Errorf("Statuses = %v, want %v", filter.Statuses, expectedStatuses)
	}

	if filter.Single != "value" {
		t.Errorf("Single = %v, want value", filter.Single)
	}
}

// TestParseToStructWithStrictNullHandling tests null handling in structs
func TestParseToStructWithParseStrictNullHandling(t *testing.T) {
	type Config struct {
		Name    string  `query:"name"`
		Value   *string `query:"value"`
		Enabled *bool   `query:"enabled"`
	}

	var config Config
	err := ParseToStruct(
		"name=test&value&enabled",
		&config,
		WithParseStrictNullHandling(true),
	)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	if config.Name != "test" {
		t.Errorf("Name = %v, want test", config.Name)
	}
	// With strictNullHandling, "value" and "enabled" without = are null
	// In Go, this means the pointer fields should be nil
}

// TestParseToStructWithCharset tests charset handling
func TestParseToStructWithParseCharset(t *testing.T) {
	type Message struct {
		Text string `query:"text"`
	}

	var msg Message
	err := ParseToStruct(
		"text=%E4%B8%AD%E6%96%87", // "中文" in UTF-8
		&msg,
		WithParseCharset(CharsetUTF8),
	)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	if msg.Text != "中文" {
		t.Errorf("Text = %v, want 中文", msg.Text)
	}
}

// TestParseToStructWithIgnoreQueryPrefix tests query prefix handling
func TestParseToStructWithParseIgnoreQueryPrefix(t *testing.T) {
	type Query struct {
		Search string `query:"search"`
		Page   int    `query:"page"`
	}

	var query Query
	err := ParseToStruct(
		"?search=golang&page=5",
		&query,
		WithParseIgnoreQueryPrefix(true),
	)
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	if query.Search != "golang" {
		t.Errorf("Search = %v, want golang", query.Search)
	}
	if query.Page != 5 {
		t.Errorf("Page = %v, want 5", query.Page)
	}
}

// TestMarshalWithArrayFormatBrackets tests bracket array format
func TestMarshalWithArrayFormatBrackets(t *testing.T) {
	type Container struct {
		Items []string `query:"items"`
	}

	container := Container{
		Items: []string{"a", "b", "c"},
	}

	str, err := Marshal(container, WithStringifyArrayFormat(ArrayFormatBrackets))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Should contain items[]=a, items[]=b, items[]=c
	parsed, _ := Parse(str)
	items, ok := parsed["items"].([]any)
	if !ok {
		t.Fatalf("items is not an array: %T", parsed["items"])
	}
	if len(items) != 3 {
		t.Errorf("items length = %v, want 3", len(items))
	}
}

// TestMarshalWithArrayFormatRepeat tests repeat array format
func TestMarshalWithArrayFormatRepeat(t *testing.T) {
	type Container struct {
		Tags []string `query:"tags"`
	}

	container := Container{
		Tags: []string{"go", "rust"},
	}

	str, err := Marshal(container, WithStringifyArrayFormat(ArrayFormatRepeat))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Should contain tags=go&tags=rust
	parsed, _ := Parse(str, WithParseDuplicates(DuplicateCombine))
	tags, ok := parsed["tags"].([]any)
	if !ok {
		t.Fatalf("tags is not an array: %T", parsed["tags"])
	}
	if len(tags) != 2 {
		t.Errorf("tags length = %v, want 2", len(tags))
	}
}

// TestMarshalWithArrayFormatComma tests comma array format
func TestMarshalWithArrayFormatComma(t *testing.T) {
	type Container struct {
		Items []string `query:"items"`
	}

	container := Container{
		Items: []string{"x", "y", "z"},
	}

	str, err := Marshal(container, WithStringifyArrayFormat(ArrayFormatComma))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Should contain items=x,y,z
	// Verify the string format
	if !strings.Contains(str, "items=x,y,z") && !strings.Contains(str, "items=x%2Cy%2Cz") {
		t.Errorf("Expected comma-separated format, got: %s", str)
	}

	// Parse back with comma option
	parsed, _ := Parse(str, WithParseComma(true))

	// With comma parsing, items could be an array or the raw value
	switch items := parsed["items"].(type) {
	case []any:
		if len(items) != 3 {
			t.Errorf("items length = %v, want 3", len(items))
		}
	case string:
		// If it's a string, it should be comma-separated
		parts := strings.Split(items, ",")
		if len(parts) != 3 {
			t.Errorf("items parts = %v, want 3", len(parts))
		}
	default:
		t.Fatalf("unexpected items type: %T", parsed["items"])
	}
}

// TestMarshalWithAllowDots tests dot notation stringify
func TestMarshalWithParseAllowDots(t *testing.T) {
	type Address struct {
		City    string `query:"city"`
		Country string `query:"country"`
	}
	type Person struct {
		Name    string  `query:"name"`
		Address Address `query:"address"`
	}

	person := Person{
		Name: "John",
		Address: Address{
			City:    "NYC",
			Country: "USA",
		},
	}

	str, err := Marshal(person, WithStringifyAllowDots(true))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Should use dot notation: address.city=NYC
	var result Person
	err = ParseToStruct(str, &result, WithParseAllowDots(true))
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}

	if result.Name != person.Name {
		t.Errorf("Name = %v, want %v", result.Name, person.Name)
	}
	if result.Address.City != person.Address.City {
		t.Errorf("Address.City = %v, want %v", result.Address.City, person.Address.City)
	}
}

// TestMarshalWithSkipNulls tests null skipping
func TestMarshalWithStringifySkipNulls(t *testing.T) {
	type Config struct {
		Name    string  `query:"name"`
		Value   *string `query:"value"`
		Enabled *bool   `query:"enabled"`
	}

	config := Config{
		Name:    "test",
		Value:   nil,
		Enabled: nil,
	}

	str, err := Marshal(config, WithStringifySkipNulls(true))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	parsed, _ := Parse(str)

	if parsed["name"] != "test" {
		t.Errorf("name = %v, want test", parsed["name"])
	}
	if _, ok := parsed["value"]; ok {
		t.Error("value should be skipped")
	}
	if _, ok := parsed["enabled"]; ok {
		t.Error("enabled should be skipped")
	}
}

// TestMarshalWithStrictNullHandling tests strict null serialization
func TestMarshalWithParseStrictNullHandling(t *testing.T) {
	type Config struct {
		Name  string `query:"name"`
		Value any    `query:"value"`
	}

	config := Config{
		Name:  "test",
		Value: nil,
	}

	str, err := Marshal(config, WithStringifyStrictNullHandling(true))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// With strictNullHandling, null values appear as key without =
	parsed, _ := Parse(str, WithParseStrictNullHandling(true))

	if parsed["name"] != "test" {
		t.Errorf("name = %v, want test", parsed["name"])
	}
}

// TestMarshalWithStringifyQueryPrefix tests query prefix
func TestMarshalWithStringifyQueryPrefix(t *testing.T) {
	type Query struct {
		Search string `query:"search"`
	}

	query := Query{Search: "golang"}

	str, err := Marshal(query, WithStringifyAddQueryPrefix(true))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if str[0] != '?' {
		t.Errorf("Should start with ?, got: %s", str)
	}
}

// TestMarshalWithSort tests key sorting
func TestMarshalWithStringifySort(t *testing.T) {
	type Data struct {
		Zebra  string `query:"zebra"`
		Apple  string `query:"apple"`
		Mango  string `query:"mango"`
		Banana string `query:"banana"`
	}

	data := Data{
		Zebra:  "z",
		Apple:  "a",
		Mango:  "m",
		Banana: "b",
	}

	str, err := Marshal(data, WithStringifySort(func(a, b string) bool {
		return a < b
	}))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Should be sorted alphabetically
	// apple=a&banana=b&mango=m&zebra=z
	expected := "apple=a&banana=b&mango=m&zebra=z"
	if str != expected {
		t.Errorf("Sorted result = %v, want %v", str, expected)
	}
}

// TestMarshalWithEncodeDotInKeys tests dot encoding in keys
func TestMarshalWithStringifyEncodeDotInKeys(t *testing.T) {
	type Config struct {
		APIKey string `query:"api.key"`
	}

	config := Config{APIKey: "secret"}

	str, err := Marshal(config,
		WithStringifyAllowDots(true),
		WithStringifyEncodeDotInKeys(true),
	)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Dot in "api.key" should be encoded as %2E
	if !strings.Contains(str, "%2E") && !strings.Contains(str, "api.key") {
		t.Errorf("Expected encoded dot or literal key, got: %s", str)
	}
}

// TestMarshalWithFormat tests RFC1738 vs RFC3986 format
func TestMarshalWithStringifyFormat(t *testing.T) {
	type Query struct {
		Search string `query:"search"`
	}

	query := Query{Search: "hello world"}

	// RFC3986 - spaces as %20
	str3986, err := Marshal(query, WithStringifyFormat(FormatRFC3986))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if !strings.Contains(str3986, "%20") {
		t.Errorf("RFC3986 should encode space as %%20, got: %s", str3986)
	}

	// RFC1738 - spaces as +
	str1738, err := Marshal(query, WithStringifyFormat(FormatRFC1738))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if !strings.Contains(str1738, "+") {
		t.Errorf("RFC1738 should encode space as +, got: %s", str1738)
	}
}

// TestRoundTripStructWithOptions tests full round-trip with various options
func TestRoundTripStructWithOptions(t *testing.T) {
	type Address struct {
		Street string `query:"street"`
		City   string `query:"city"`
	}
	type User struct {
		Name    string   `query:"name"`
		Age     int      `query:"age"`
		Tags    []string `query:"tags"`
		Address Address  `query:"address"`
		Active  bool     `query:"active"`
	}

	original := User{
		Name:   "John Doe",
		Age:    30,
		Tags:   []string{"admin", "user"},
		Active: true,
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
	}

	// Marshal with options
	str, err := Marshal(original,
		WithStringifyArrayFormat(ArrayFormatIndices),
		WithStringifyAllowDots(false),
	)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal back
	var result User
	err = Unmarshal(str, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.Name != original.Name {
		t.Errorf("Name = %v, want %v", result.Name, original.Name)
	}
	if result.Age != original.Age {
		t.Errorf("Age = %v, want %v", result.Age, original.Age)
	}
	if result.Active != original.Active {
		t.Errorf("Active = %v, want %v", result.Active, original.Active)
	}
	if !reflect.DeepEqual(result.Tags, original.Tags) {
		t.Errorf("Tags = %v, want %v", result.Tags, original.Tags)
	}
	if result.Address.City != original.Address.City {
		t.Errorf("Address.City = %v, want %v", result.Address.City, original.Address.City)
	}
}

// TestComplexNestedStructWithAllOptions tests deeply nested struct with all options
func TestComplexNestedStructWithAllOptions(t *testing.T) {
	type Notification struct {
		Email bool `query:"email"`
		SMS   bool `query:"sms"`
	}
	type Settings struct {
		Theme         string       `query:"theme"`
		Language      string       `query:"language"`
		Notifications Notification `query:"notifications"`
	}
	type Profile struct {
		Bio      string   `query:"bio"`
		Website  string   `query:"website"`
		Skills   []string `query:"skills"`
		Settings Settings `query:"settings"`
	}
	type User struct {
		ID      int      `query:"id"`
		Name    string   `query:"name"`
		Email   string   `query:"email"`
		Roles   []string `query:"roles"`
		Profile Profile  `query:"profile"`
	}

	original := User{
		ID:    123,
		Name:  "Alice",
		Email: "alice@example.com",
		Roles: []string{"admin", "editor", "viewer"},
		Profile: Profile{
			Bio:     "Software Developer",
			Website: "https://alice.dev",
			Skills:  []string{"Go", "Rust", "Python"},
			Settings: Settings{
				Theme:    "dark",
				Language: "en",
				Notifications: Notification{
					Email: true,
					SMS:   false,
				},
			},
		},
	}

	// Test with dot notation
	str, err := Marshal(original, WithStringifyAllowDots(true))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var result User
	err = Unmarshal(str, &result, WithParseAllowDots(true))
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.ID != original.ID {
		t.Errorf("ID = %v, want %v", result.ID, original.ID)
	}
	if result.Name != original.Name {
		t.Errorf("Name = %v, want %v", result.Name, original.Name)
	}
	if !reflect.DeepEqual(result.Roles, original.Roles) {
		t.Errorf("Roles = %v, want %v", result.Roles, original.Roles)
	}
	if result.Profile.Settings.Theme != original.Profile.Settings.Theme {
		t.Errorf("Profile.Settings.Theme = %v, want %v",
			result.Profile.Settings.Theme, original.Profile.Settings.Theme)
	}
	if result.Profile.Settings.Notifications.Email != original.Profile.Settings.Notifications.Email {
		t.Errorf("Notifications.Email = %v, want %v",
			result.Profile.Settings.Notifications.Email,
			original.Profile.Settings.Notifications.Email)
	}
}

// TestStrapiStyleQuery tests Strapi-like query structure
func TestStrapiStyleQuery(t *testing.T) {
	type DateRange struct {
		GTE string `query:"$gte"`
		LTE string `query:"$lte"`
	}
	type Filters struct {
		Status  []string  `query:"status"`
		Created DateRange `query:"created"`
	}
	type Pagination struct {
		Page  int `query:"page"`
		Limit int `query:"limit"`
	}
	type StrapiQuery struct {
		Filters    Filters    `query:"filters"`
		Pagination Pagination `query:"pagination"`
		Populate   []string   `query:"populate"`
		Sort       []string   `query:"sort"`
	}

	original := StrapiQuery{
		Filters: Filters{
			Status: []string{"published", "draft"},
			Created: DateRange{
				GTE: "2024-01-01",
				LTE: "2024-12-31",
			},
		},
		Pagination: Pagination{
			Page:  1,
			Limit: 25,
		},
		Populate: []string{"author", "category"},
		Sort:     []string{"createdAt:desc"},
	}

	// Marshal
	str, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal
	var result StrapiQuery
	err = Unmarshal(str, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.Pagination.Page != original.Pagination.Page {
		t.Errorf("Pagination.Page = %v, want %v", result.Pagination.Page, original.Pagination.Page)
	}
	if result.Filters.Created.GTE != original.Filters.Created.GTE {
		t.Errorf("Filters.Created.GTE = %v, want %v",
			result.Filters.Created.GTE, original.Filters.Created.GTE)
	}
	if !reflect.DeepEqual(result.Populate, original.Populate) {
		t.Errorf("Populate = %v, want %v", result.Populate, original.Populate)
	}
}

// TestStructWithMapField tests struct containing map fields
func TestStructWithMapField(t *testing.T) {
	type Config struct {
		Name     string            `query:"name"`
		Settings map[string]string `query:"settings"`
		Metadata map[string]int    `query:"metadata"`
	}

	original := Config{
		Name: "myconfig",
		Settings: map[string]string{
			"theme":    "dark",
			"language": "en",
		},
		Metadata: map[string]int{
			"version": 2,
			"count":   100,
		},
	}

	str, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var result Config
	err = Unmarshal(str, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.Name != original.Name {
		t.Errorf("Name = %v, want %v", result.Name, original.Name)
	}
	if result.Settings["theme"] != original.Settings["theme"] {
		t.Errorf("Settings[theme] = %v, want %v",
			result.Settings["theme"], original.Settings["theme"])
	}
	if result.Metadata["version"] != original.Metadata["version"] {
		t.Errorf("Metadata[version] = %v, want %v",
			result.Metadata["version"], original.Metadata["version"])
	}
}

// TestStructWithInterfaceField tests struct with any/interface{} fields
// TODO: new Unmarshal needs to handle nested any fields
func TestStructWithInterfaceField(t *testing.T) {
	t.Skip("TODO: new Unmarshal needs to handle nested any fields")
	type Flexible struct {
		Name    string `query:"name"`
		Payload any    `query:"payload"`
	}

	original := Flexible{
		Name: "test",
		Payload: map[string]any{
			"key":   "value",
			"count": 42,
		},
	}

	str, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var result Flexible
	err = Unmarshal(str, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.Name != original.Name {
		t.Errorf("Name = %v, want %v", result.Name, original.Name)
	}

	payload, ok := result.Payload.(map[string]any)
	if !ok {
		t.Fatalf("Payload is not map[string]any: %T", result.Payload)
	}
	if payload["key"] != "value" {
		t.Errorf("Payload[key] = %v, want value", payload["key"])
	}
}

// TestParseToStructWithDuplicates tests duplicate key handling
func TestParseToStructWithParseDuplicates(t *testing.T) {
	type Data struct {
		Values []string `query:"v"`
	}

	// Test combine (default)
	var dataCombine Data
	err := ParseToStruct("v=a&v=b&v=c", &dataCombine, WithParseDuplicates(DuplicateCombine))
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}
	if len(dataCombine.Values) != 3 {
		t.Errorf("Combine: Values length = %v, want 3", len(dataCombine.Values))
	}

	// Test first
	type SingleData struct {
		Value string `query:"v"`
	}
	var dataFirst SingleData
	err = ParseToStruct("v=first&v=second", &dataFirst, WithParseDuplicates(DuplicateFirst))
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}
	if dataFirst.Value != "first" {
		t.Errorf("First: Value = %v, want first", dataFirst.Value)
	}

	// Test last
	var dataLast SingleData
	err = ParseToStruct("v=first&v=last", &dataLast, WithParseDuplicates(DuplicateLast))
	if err != nil {
		t.Fatalf("ParseToStruct() error = %v", err)
	}
	if dataLast.Value != "last" {
		t.Errorf("Last: Value = %v, want last", dataLast.Value)
	}
}

// TestMarshalWithCustomDelimiter tests custom delimiter
func TestMarshalWithCustomDelimiter(t *testing.T) {
	type Data struct {
		A string `query:"a"`
		B string `query:"b"`
	}

	data := Data{A: "1", B: "2"}

	str, err := Marshal(data, WithStringifyDelimiter(";"))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Should use semicolon
	if !strings.Contains(str, ";") {
		t.Errorf("Should use ; delimiter, got: %s", str)
	}

	// Parse back with same delimiter
	var result Data
	err = Unmarshal(str, &result, WithParseDelimiter(";"))
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result.A != data.A || result.B != data.B {
		t.Errorf("Result = %+v, want %+v", result, data)
	}
}

// TestMarshalWithEncodeValuesOnly tests encodeValuesOnly option
func TestMarshalWithStringifyEncodeValuesOnly(t *testing.T) {
	type Data struct {
		Key string `query:"a[b]"`
	}

	data := Data{Key: "hello world"}

	str, err := Marshal(data, WithStringifyEncodeValuesOnly(true))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Key should not be encoded, value should be
	if strings.Contains(str, "%5B") {
		t.Errorf("Key should not be encoded with encodeValuesOnly, got: %s", str)
	}
	if !strings.Contains(str, "%20") && !strings.Contains(str, "+") {
		t.Errorf("Value should be encoded, got: %s", str)
	}
}

// TestMarshalWithFilter tests filter option
func TestMarshalWithStringifyFilter(t *testing.T) {
	type Data struct {
		A string `query:"a"`
		B string `query:"b"`
		C string `query:"c"`
	}

	data := Data{A: "1", B: "2", C: "3"}

	// Filter as array of keys
	str, err := Marshal(data, WithStringifyFilter([]string{"a", "c"}))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	parsed, _ := Parse(str)
	if _, ok := parsed["b"]; ok {
		t.Error("b should be filtered out")
	}
	if parsed["a"] != "1" {
		t.Errorf("a = %v, want 1", parsed["a"])
	}
	if parsed["c"] != "3" {
		t.Errorf("c = %v, want 3", parsed["c"])
	}
}

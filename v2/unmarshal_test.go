// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"reflect"
	"testing"
)

func TestUnmarshalBytes_SimpleStruct(t *testing.T) {
	type User struct {
		Name  string `query:"name"`
		Age   int    `query:"age"`
		Email string `query:"email"`
	}

	var user User
	err := UnmarshalBytes([]byte("name=John&age=30&email=john@example.com"), &user)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Name: got %q, want %q", user.Name, "John")
	}
	if user.Age != 30 {
		t.Errorf("Age: got %d, want %d", user.Age, 30)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Email: got %q, want %q", user.Email, "john@example.com")
	}
}

func TestUnmarshalBytes_NestedStruct(t *testing.T) {
	type Profile struct {
		Name string `query:"name"`
		Age  string `query:"age"`
	}
	type Settings struct {
		Theme string `query:"theme"`
		Lang  string `query:"lang"`
	}
	type User struct {
		Profile  Profile  `query:"profile"`
		Settings Settings `query:"settings"`
	}
	type Data struct {
		User User `query:"user"`
	}

	var data Data
	err := UnmarshalBytes([]byte("user[profile][name]=John&user[profile][age]=30&user[settings][theme]=dark&user[settings][lang]=en"), &data)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	if data.User.Profile.Name != "John" {
		t.Errorf("Profile.Name: got %q, want %q", data.User.Profile.Name, "John")
	}
	if data.User.Settings.Theme != "dark" {
		t.Errorf("Settings.Theme: got %q, want %q", data.User.Settings.Theme, "dark")
	}
}

func TestUnmarshalBytes_ArrayStruct(t *testing.T) {
	type Data struct {
		Items []string `query:"items"`
	}

	var data Data
	err := UnmarshalBytes([]byte("items[0]=apple&items[1]=banana&items[2]=cherry"), &data)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	expected := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(data.Items, expected) {
		t.Errorf("Items: got %v, want %v", data.Items, expected)
	}
}

func TestUnmarshalBytes_Map(t *testing.T) {
	var data map[string]any
	err := UnmarshalBytes([]byte("name=John&age=30"), &data)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	if data["name"] != "John" {
		t.Errorf("name: got %v, want %q", data["name"], "John")
	}
	if data["age"] != "30" {
		t.Errorf("age: got %v, want %q", data["age"], "30")
	}
}

func TestUnmarshalBytes_NestedMap(t *testing.T) {
	var data map[string]any
	err := UnmarshalBytes([]byte("user[name]=John&user[age]=30"), &data)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	user, ok := data["user"].(map[string]any)
	if !ok {
		t.Fatalf("user is not a map: %T", data["user"])
	}
	if user["name"] != "John" {
		t.Errorf("user.name: got %v, want %q", user["name"], "John")
	}
}

func TestUnmarshalBytes_Interface(t *testing.T) {
	var data any
	err := UnmarshalBytes([]byte("name=John&age=30"), &data)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("data is not a map: %T", data)
	}
	if m["name"] != "John" {
		t.Errorf("name: got %v, want %q", m["name"], "John")
	}
}

func TestUnmarshalBytes_WithOptions(t *testing.T) {
	type Data struct {
		A struct {
			B struct {
				C string `query:"c"`
			} `query:"b"`
		} `query:"a"`
	}

	var data Data
	err := UnmarshalBytes([]byte("a.b.c=value"), &data, WithParseAllowDots(true))
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	if data.A.B.C != "value" {
		t.Errorf("A.B.C: got %q, want %q", data.A.B.C, "value")
	}
}

func TestUnmarshalBytes_MixedStructWithMap(t *testing.T) {
	type Request struct {
		Page    int            `query:"page"`
		Filters map[string]any `query:"filters"`
	}

	var req Request
	err := UnmarshalBytes([]byte("page=1&filters[status]=active&filters[category]=tech"), &req)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	if req.Page != 1 {
		t.Errorf("Page: got %d, want %d", req.Page, 1)
	}
	if req.Filters["status"] != "active" {
		t.Errorf("Filters.status: got %v, want %q", req.Filters["status"], "active")
	}
	if req.Filters["category"] != "tech" {
		t.Errorf("Filters.category: got %v, want %q", req.Filters["category"], "tech")
	}
}

func TestUnmarshalBytes_Pointer(t *testing.T) {
	type User struct {
		Name *string `query:"name"`
		Age  *int    `query:"age"`
	}

	var user User
	err := UnmarshalBytes([]byte("name=John&age=30"), &user)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	if user.Name == nil || *user.Name != "John" {
		t.Errorf("Name: got %v, want %q", user.Name, "John")
	}
	if user.Age == nil || *user.Age != 30 {
		t.Errorf("Age: got %v, want %d", user.Age, 30)
	}
}

func TestUnmarshalBytes_URLEncoded(t *testing.T) {
	type User struct {
		Name  string `query:"name"`
		Email string `query:"email"`
	}

	var user User
	err := UnmarshalBytes([]byte("name=John%20Doe&email=john%40example.com"), &user)
	if err != nil {
		t.Fatalf("UnmarshalBytes: %v", err)
	}

	if user.Name != "John Doe" {
		t.Errorf("Name: got %q, want %q", user.Name, "John Doe")
	}
	if user.Email != "john@example.com" {
		t.Errorf("Email: got %q, want %q", user.Email, "john@example.com")
	}
}

func TestUnmarshalString(t *testing.T) {
	type User struct {
		Name string `query:"name"`
	}

	var user User
	err := UnmarshalString("name=John", &user)
	if err != nil {
		t.Fatalf("UnmarshalString: %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Name: got %q, want %q", user.Name, "John")
	}
}

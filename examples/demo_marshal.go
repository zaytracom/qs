package main

import (
	"fmt"

	qs "github.com/zaytracom/qs/v1"
)

type Person struct {
	Name     string  `query:"name"`
	Age      int     `query:"age"`
	Email    string  `query:"email"`
	IsActive bool    `query:"active"`
	Score    float64 `query:"score"`
}

type Filter struct {
	Query    string   `query:"q"`
	Tags     []string `query:"tags"`
	Category string   `query:"category"`
	MinPrice int      `query:"min_price"`
	MaxPrice int      `query:"max_price"`
}

// DemoMarshalUnmarshal demonstrates the new idiomatic Marshal/Unmarshal functions
func DemoMarshalUnmarshal() {
	fmt.Println("=== QS Marshal/Unmarshal Demo ===")

	// Demo 1: Automatic type detection with structs
	fmt.Println("1. Struct Marshal/Unmarshal:")
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		IsActive: true,
		Score:    95.5,
	}

	// Marshal struct -> query string (automatic type detection)
	queryString, err := qs.Marshal(person)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Person: %+v\n", person)
	fmt.Printf("Query: %s\n", queryString)

	// Unmarshal query string -> struct (automatic type detection)
	var newPerson Person
	err = qs.Unmarshal(queryString, &newPerson)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parsed: %+v\n", newPerson)
	fmt.Printf("Equal: %t\n\n", person == newPerson)

	// Demo 2: Automatic type detection with maps
	fmt.Println("2. Map Marshal/Unmarshal:")
	data := map[string]interface{}{
		"product": "laptop",
		"brand":   "Apple",
		"price":   2500,
		"tags":    []interface{}{"electronics", "computer", "apple"},
	}

	// Marshal map -> query string (automatic type detection)
	mapQuery, err := qs.Marshal(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Map: %+v\n", data)
	fmt.Printf("Query: %s\n", mapQuery)

	// Unmarshal query string -> map (automatic type detection)
	var parsedData map[string]interface{}
	err = qs.Unmarshal(mapQuery, &parsedData)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parsed: %+v\n\n", parsedData)

	// Demo 3: Runtime type detection - same API for different types
	fmt.Println("3. Runtime type detection:")
	testQuery := "name=TestUser&age=25&q=search&tags[]=tag1&tags[]=tag2&category=test&min_price=10&max_price=100"
	fmt.Printf("Test query: %s\n", testQuery)

	// Try parsing into different types with the same function
	targets := []interface{}{
		&Person{},                 // struct
		&map[string]interface{}{}, // map
		&Filter{},                 // different struct
	}

	for i, target := range targets {
		err := qs.Unmarshal(testQuery, target)
		if err != nil {
			fmt.Printf("Target %d (%T) error: %v\n", i+1, target, err)
		} else {
			fmt.Printf("Target %d (%T): %+v\n", i+1, target, target)
		}
	}

	fmt.Println("\n4. HTTP-like usage example:")
	simulateHTTPHandler("product=laptop&brand=Apple&price=1500&features[]=retina&features[]=m2&metadata[weight]=1.4kg&metadata[color]=silver")
}

func simulateHTTPHandler(queryString string) {
	fmt.Printf("Incoming query: %s\n", queryString)

	// Option 1: Parse into a specific struct (if you know the structure)
	type ProductQuery struct {
		Product  string            `query:"product"`
		Brand    string            `query:"brand"`
		Price    int               `query:"price"`
		Features []string          `query:"features"`
		Metadata map[string]string `query:"metadata"`
	}

	var productQuery ProductQuery
	if err := qs.Unmarshal(queryString, &productQuery); err == nil {
		fmt.Printf("As ProductQuery: %+v\n", productQuery)
	}

	// Option 2: Parse into a generic map (for dynamic handling)
	var genericParams map[string]interface{}
	if err := qs.Unmarshal(queryString, &genericParams); err == nil {
		fmt.Printf("As generic map: %+v\n", genericParams)
	}

	// Option 3: Create response and marshal it back
	response := map[string]interface{}{
		"status":    "success",
		"query":     queryString,
		"parsed":    productQuery,
		"timestamp": "2023-11-15T10:30:00Z",
	}

	responseQuery, _ := qs.Marshal(response)
	fmt.Printf("Response as query: %s\n", responseQuery)
}

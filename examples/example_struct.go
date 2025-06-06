package main

import (
	"fmt"
	"strings"

	"github.com/zaytracom/qs/v1"
)

type User struct {
	Name     string  `query:"name"`
	Age      int     `query:"age"`
	Email    string  `query:"email"`
	IsActive bool    `query:"active"`
	Score    float64 `query:"score"`
}

type SearchFilter struct {
	Query    string   `query:"q"`
	Tags     []string `query:"tags"`
	Category string   `query:"category"`
	MinPrice int      `query:"min_price"`
	MaxPrice int      `query:"max_price"`
}

func main() {
	fmt.Println("=== QS Struct Parsing Examples ===")

	// Example 1: Basic struct parsing
	fmt.Println("1. Basic struct parsing:")
	queryString := "name=John&age=30&email=john@example.com&active=true&score=95.5"
	fmt.Printf("Query: %s\n", queryString)

	var user User
	err := qs.ParseToStruct(queryString, &user)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Parsed User: %+v\n\n", user)

	// Example 2: Struct with arrays
	fmt.Println("2. Struct with arrays:")
	filterQuery := "q=golang&tags[]=programming&tags[]=web&category=tech&min_price=10&max_price=100"
	fmt.Printf("Query: %s\n", filterQuery)

	var filter SearchFilter
	err = qs.ParseToStruct(filterQuery, &filter)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Parsed Filter: %+v\n\n", filter)

	// Example 3: Converting struct back to query string
	fmt.Println("3. Converting struct to query string:")
	newUser := &User{
		Name:     "Alice",
		Age:      25,
		Email:    "alice@example.com",
		IsActive: false,
		Score:    88.5,
	}

	fmt.Printf("User struct: %+v\n", newUser)

	queryString, err = qs.StructToQueryString(newUser)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated query: %s\n\n", queryString)

	// Example 4: Round trip (struct -> query -> struct)
	fmt.Println("4. Round trip test:")
	var roundTripUser User
	err = qs.ParseToStruct(queryString, &roundTripUser)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Original: %+v\n", newUser)
	fmt.Printf("Round trip: %+v\n", roundTripUser)
	fmt.Printf("Equal: %t\n", *newUser == roundTripUser)

	fmt.Println("\n=== New Marshal/Unmarshal API Demo ===")

	// Example 5: Using new idiomatic Marshal/Unmarshal functions
	fmt.Println("5. Idiomatic Marshal/Unmarshal (runtime type detection):")

	// Same struct, but using the new API
	person := User{
		Name:     "Bob",
		Age:      35,
		Email:    "bob@test.com",
		IsActive: true,
		Score:    92.0,
	}

	// Marshal using new API (automatic type detection)
	marshaledQuery, err := qs.Marshal(person)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Marshal result: %s\n", marshaledQuery)

	// Unmarshal using new API (automatic type detection)
	var unmarshaledPerson User
	err = qs.Unmarshal(marshaledQuery, &unmarshaledPerson)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Unmarshal result: %+v\n", unmarshaledPerson)

	// Example 6: Runtime type detection with different types
	fmt.Println("\n6. Runtime type detection - same API for different types:")
	testQuery := "name=Alice&age=28&q=search&tags[]=go&tags[]=api&category=backend"

	// Using the same Unmarshal function for different types
	var userTarget User
	var filterTarget SearchFilter
	var mapTarget map[string]interface{}

	fmt.Printf("Test query: %s\n", testQuery)

	if err := qs.Unmarshal(testQuery, &userTarget); err == nil {
		fmt.Printf("As User: %+v\n", userTarget)
	}

	if err := qs.Unmarshal(testQuery, &filterTarget); err == nil {
		fmt.Printf("As SearchFilter: %+v\n", filterTarget)
	}

	if err := qs.Unmarshal(testQuery, &mapTarget); err == nil {
		fmt.Printf("As Map: %+v\n", mapTarget)
	}

	// Run extended Strapi-like demonstration
	fmt.Println("\n" + strings.Repeat("=", 50))
	RunStrapiDemo()
}

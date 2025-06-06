# QS Go Library - Usage Examples

The QS library for Go provides powerful capabilities for working with query strings, including parsing and creating complex nested data structures.

## Idiomatic Marshal/Unmarshal Functions

### Automatic type detection at runtime

The new `Marshal` and `Unmarshal` functions automatically detect data types at runtime and choose the appropriate processing method.

```go
package main

import (
    "fmt"
    "github.com/zaytra/qs/qs"
)

type User struct {
    Name     string  `query:"name"`
    Age      int     `query:"age"`
    Email    string  `query:"email"`
    IsActive bool    `query:"active"`
    Score    float64 `query:"score"`
}

func main() {
    // Marshal struct -> query string (automatic type detection)
    user := User{
        Name:     "John Doe",
        Age:      30,
        Email:    "john@example.com",
        IsActive: true,
        Score:    95.5,
    }

    queryString, err := qs.Marshal(user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Query: %s\n", queryString)
    // Output: name=John%20Doe&age=30&email=john%40example.com&active=true&score=95.5

    // Unmarshal query string -> struct (automatic type detection)
    var newUser User
    err = qs.Unmarshal(queryString, &newUser)
    if err != nil {
        panic(err)
    }
    fmt.Printf("User: %+v\n", newUser)
    // Output: {Name:John Doe Age:30 Email:john@example.com IsActive:true Score:95.5}
}
```

### Working with different data types using single API

```go
func runtimeTypeDetection() {
    testQuery := "name=Alice&age=28&q=search&tags[]=go&tags[]=api&category=backend"

    // Same API for different types
    var userTarget User
    var mapTarget map[string]interface{}
    var filterTarget SearchFilter

    // Parse to struct
    if err := qs.Unmarshal(testQuery, &userTarget); err == nil {
        fmt.Printf("As User: %+v\n", userTarget)
    }

    // Parse to map
    if err := qs.Unmarshal(testQuery, &mapTarget); err == nil {
        fmt.Printf("As Map: %+v\n", mapTarget)
    }

    // Parse to another struct
    if err := qs.Unmarshal(testQuery, &filterTarget); err == nil {
        fmt.Printf("As Filter: %+v\n", filterTarget)
    }
}
```

### HTTP Handler example

```go
func httpHandlerExample(queryString string) {
    // Option 1: Parse to specific structure
    type ProductQuery struct {
        Product  string            `query:"product"`
        Brand    string            `query:"brand"`
        Price    int               `query:"price"`
        Features []string          `query:"features"`
        Metadata map[string]string `query:"metadata"`
    }

    var productQuery ProductQuery
    if err := qs.Unmarshal(queryString, &productQuery); err == nil {
        fmt.Printf("Product Query: %+v\n", productQuery)
    }

    // Option 2: Parse to generic map for dynamic processing
    var genericParams map[string]interface{}
    if err := qs.Unmarshal(queryString, &genericParams); err == nil {
        fmt.Printf("Generic Params: %+v\n", genericParams)
    }

    // Create response and marshal back
    response := map[string]interface{}{
        "status":    "success",
        "query":     queryString,
        "parsed":    productQuery,
        "timestamp": "2023-11-15T10:30:00Z",
    }

    responseQuery, _ := qs.Marshal(response)
    fmt.Printf("Response Query: %s\n", responseQuery)
}
```

### Complex nested structures with Marshal/Unmarshal

```go
type ComplexData struct {
    User     User              `query:"user"`
    Filters  SearchFilter      `query:"filters"`
    Settings map[string]string `query:"settings"`
    Enabled  bool              `query:"enabled"`
}

func complexStructExample() {
    complex := ComplexData{
        User: User{
            Name:     "Alice",
            Age:      25,
            Email:    "alice@example.com",
            IsActive: false,
            Score:    88.0,
        },
        Filters: SearchFilter{
            Query:    "golang developer",
            Tags:     []string{"programming", "backend", "api"},
            Category: "tech",
            MinPrice: 50000,
            MaxPrice: 150000,
        },
        Settings: map[string]string{
            "theme":    "dark",
            "language": "en",
            "timezone": "UTC",
        },
        Enabled: true,
    }

    // Marshal complex structure
    queryString, err := qs.Marshal(complex)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Complex Query: %s\n", queryString)

    // Unmarshal back
    var parsedComplex ComplexData
    err = qs.Unmarshal(queryString, &parsedComplex)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Parsed Complex: %+v\n", parsedComplex)
}
```

## Basic Examples

### Simple parsing

```go
package main

import (
    "fmt"
    "github.com/zaytra/qs/qs"
)

func main() {
    // Simple query string
    result, err := qs.Parse("name=John&age=30&city=NYC")
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", result)
    // Output: map[age:30 city:NYC name:John]
}
```

### Simple query string creation

```go
func main() {
    data := map[string]interface{}{
        "name": "John",
        "age":  30,
        "city": "NYC",
    }

    result, err := qs.Stringify(data)
    if err != nil {
        panic(err)
    }
    fmt.Println(result)
    // Output: age=30&city=NYC&name=John
}
```

## Complex nested structures

### 1. User profile

```go
func userProfileExample() {
    // Deeply nested user structure
    queryString := "user[profile][personal][name]=John&user[profile][personal][surname]=Smith&user[profile][settings][theme][mode]=dark&user[profile][settings][theme][colors][primary]=%23007bff&user[profile][settings][notifications][email]=true&user[profile][settings][notifications][push]=false&user[profile][settings][language]=en"

    result, err := qs.Parse(queryString)
    if err != nil {
        panic(err)
    }

    // Result:
    // map[
    //   user:map[
    //     profile:map[
    //       personal:map[
    //         name:John
    //         surname:Smith
    //       ]
    //       settings:map[
    //         theme:map[
    //           mode:dark
    //           colors:map[primary:#007bff]
    //         ]
    //         notifications:map[email:true push:false]
    //         language:en
    //       ]
    //     ]
    //   ]
    // ]
}
```

### 2. E-commerce product catalog

```go
func ecommerceExample() {
    // Complex structure for online store
    queryString := `products[0][id]=123&products[0][name]=MacBook Pro&products[0][price]=2499&products[0][currency]=USD&products[0][category][main]=computers&products[0][category][sub]=laptops&products[0][specifications][cpu]=M2 Pro&products[0][specifications][memory]=16GB&products[0][specifications][storage]=512GB&products[0][images][]=https://example.com/img1.jpg&products[0][images][]=https://example.com/img2.jpg&products[0][variants][0][color]=silver&products[0][variants][0][stock]=5&products[0][variants][1][color]=space-gray&products[0][variants][1][stock]=3&products[1][id]=456&products[1][name]=iPhone 14&products[1][price]=999&filters[price][min]=500&filters[price][max]=3000&filters[categories][]=computers&filters[categories][]=phones&sort[field]=price&sort[direction]=asc&pagination[page]=1&pagination[limit]=20`

    result, err := qs.Parse(queryString)
    if err != nil {
        panic(err)
    }

    // Create back to query string
    queryBack, err := qs.Stringify(result)
    if err != nil {
        panic(err)
    }

    fmt.Println("Restored query string:", queryBack)
}
```

### 3. API request with filters and nested relationships

```go
func apiRequestExample() {
    // JSON API style request
    queryString := `include[]=author&include[]=comments.user&fields[posts]=title,content,published_at&fields[users]=name,email&filter[published]=true&filter[author][role]=admin&filter[created_at][from]=2023-01-01&filter[created_at][to]=2023-12-31&sort[]=created_at&sort[]=-author.name&page[number]=1&page[size]=10&extra[meta][analytics]=true&extra[debug][sql]=false`

    result, err := qs.Parse(queryString)
    if err != nil {
        panic(err)
    }

    // Result allows easy data extraction:
    if filters, ok := result["filter"].(map[string]interface{}); ok {
        if author, ok := filters["author"].(map[string]interface{}); ok {
            role := author["role"]
            fmt.Printf("Filter by author role: %s\n", role)
        }
    }
}
```

### 4. Form with file upload and validation

```go
func formWithFilesExample() {
    queryString := `form[title]=User Registration Form&form[description]=Please fill all fields&form[fields][0][type]=text&form[fields][0][name]=firstName&form[fields][0][label]=First Name&form[fields][0][required]=true&form[fields][0][validation][minLength]=2&form[fields][0][validation][maxLength]=50&form[fields][0][validation][pattern]=^[A-Za-z\s]+$&form[fields][1][type]=email&form[fields][1][name]=email&form[fields][1][label]=Email&form[fields][1][required]=true&form[fields][1][validation][pattern]=^[^\s@]+@[^\s@]+\.[^\s@]+$&form[fields][2][type]=file&form[fields][2][name]=avatar&form[fields][2][label]=Avatar&form[fields][2][accept][]=image/jpeg&form[fields][2][accept][]=image/png&form[fields][2][maxSize]=5242880&form[files][uploaded][0][name]=photo.jpg&form[files][uploaded][0][size]=1024768&form[files][uploaded][0][type]=image/jpeg&form[files][uploaded][0][url]=https://storage.example.com/files/photo.jpg&form[metadata][created]=2023-11-15T10:30:00Z&form[metadata][author][id]=42&form[metadata][author][name]=Admin&form[metadata][version]=1.2.3`

    result, err := qs.Parse(queryString)
    if err != nil {
        panic(err)
    }

    // Extract form information
    if form, ok := result["form"].(map[string]interface{}); ok {
        title := form["title"]
        fmt.Printf("Form title: %s\n", title)

        if fields, ok := form["fields"].(map[string]interface{}); ok {
            fmt.Printf("Number of fields: %d\n", len(fields))
        }
    }
}
```

### 5. Complex analytics and metrics

```go
func analyticsExample() {
    queryString := `analytics[metrics][pageviews][total]=15432&analytics[metrics][pageviews][unique]=8721&analytics[metrics][sessions][total]=5234&analytics[metrics][sessions][bounce_rate]=0.35&analytics[dimensions][country][US]=8500&analytics[dimensions][country][RU]=3200&analytics[dimensions][country][DE]=1800&analytics[dimensions][device][desktop]=9800&analytics[dimensions][device][mobile]=4500&analytics[dimensions][device][tablet]=1132&analytics[segments][0][name]=returning_users&analytics[segments][0][value]=3456&analytics[segments][1][name]=new_users&analytics[segments][1][value]=1778&analytics[time_range][start]=2023-11-01T00:00:00Z&analytics[time_range][end]=2023-11-30T23:59:59Z&analytics[aggregation][interval]=day&analytics[aggregation][timezone]=UTC&analytics[filters][include][pages][]=\/&analytics[filters][include][pages][]=%2Fabout&analytics[filters][exclude][bots]=true&analytics[export][format]=json&analytics[export][fields][]=pageviews&analytics[export][fields][]=sessions`

    result, err := qs.Parse(queryString)
    if err != nil {
        panic(err)
    }

    // Work with analytical data
    if analytics, ok := result["analytics"].(map[string]interface{}); ok {
        if metrics, ok := analytics["metrics"].(map[string]interface{}); ok {
            if pageviews, ok := metrics["pageviews"].(map[string]interface{}); ok {
                total := pageviews["total"]
                unique := pageviews["unique"]
                fmt.Printf("Page views: %s total, %s unique\n", total, unique)
            }
        }
    }
}
```

## Working with options

### Parsing with different options

```go
func parseWithOptions() {
    // With query prefix
    result1, _ := qs.Parse("?name=John&age=30", &qs.ParseOptions{
        IgnoreQueryPrefix: true,
    })

    // With custom delimiter
    result2, _ := qs.Parse("name=John;age=30;city=NYC", &qs.ParseOptions{
        Delimiter: ";",
    })

    // Strict null handling
    result3, _ := qs.Parse("name=John&empty&null=", &qs.ParseOptions{
        StrictNullHandling: true,
    })

    // With dot notation
    result4, _ := qs.Parse("user.profile.name=John&user.profile.age=30", &qs.ParseOptions{
        AllowDots: true,
    })

    fmt.Printf("Result 1: %+v\n", result1)
    fmt.Printf("Result 2: %+v\n", result2)
    fmt.Printf("Result 3: %+v\n", result3)
    fmt.Printf("Result 4: %+v\n", result4)
}
```

### Creating strings with different formats

```go
func stringifyWithOptions() {
    data := map[string]interface{}{
        "items": []interface{}{"apple", "banana", "cherry"},
        "user": map[string]interface{}{
            "name": "John",
            "age":  30,
        },
    }

    // Standard format (indices)
    result1, _ := qs.Stringify(data)

    // Brackets format for arrays
    result2, _ := qs.Stringify(data, &qs.StringifyOptions{
        ArrayFormat: "brackets",
    })

    // Repeat format for arrays
    result3, _ := qs.Stringify(data, &qs.StringifyOptions{
        ArrayFormat: "repeat",
    })

    // With query prefix
    result4, _ := qs.Stringify(data, &qs.StringifyOptions{
        AddQueryPrefix: true,
    })

    // With custom delimiter
    result5, _ := qs.Stringify(data, &qs.StringifyOptions{
        Delimiter: ";",
    })

    fmt.Printf("Indices: %s\n", result1)
    fmt.Printf("Brackets: %s\n", result2)
    fmt.Printf("Repeat: %s\n", result3)
    fmt.Printf("With prefix: %s\n", result4)
    fmt.Printf("Custom delimiter: %s\n", result5)
}
```

## Real-world scenarios

### 1. REST API filtering

```go
func restAPIExample() {
    // Build URL for API request
    filters := map[string]interface{}{
        "posts": map[string]interface{}{
            "status":    "published",
            "author_id": []interface{}{1, 2, 3},
            "tags":      []interface{}{"golang", "programming"},
            "created_at": map[string]interface{}{
                "gte": "2023-01-01",
                "lte": "2023-12-31",
            },
        },
        "include": []interface{}{"author", "comments", "tags"},
        "sort":    []interface{}{"-created_at", "title"},
        "page": map[string]interface{}{
            "number": 1,
            "size":   20,
        },
    }

    queryString, _ := qs.Stringify(filters)
    url := "https://api.example.com/posts?" + queryString

    fmt.Printf("API URL: %s\n", url)
}
```

### 2. Search form

```go
func searchFormExample() {
    // Complex search form
    searchParams := map[string]interface{}{
        "query": "golang developer",
        "location": map[string]interface{}{
            "city":    "Moscow",
            "country": "RU",
            "remote":  true,
        },
        "salary": map[string]interface{}{
            "min":      100000,
            "max":      300000,
            "currency": "RUB",
        },
        "experience": map[string]interface{}{
            "min_years": 3,
            "max_years": 7,
        },
        "skills":     []interface{}{"go", "postgresql", "docker", "kubernetes"},
        "employment": []interface{}{"full-time", "contract"},
        "sort": map[string]interface{}{
            "field": "salary",
            "order": "desc",
        },
    }

    queryString, _ := qs.Stringify(searchParams)
    searchURL := "https://jobs.example.com/search?" + queryString

    fmt.Printf("Search URL: %s\n", searchURL)
}
```

### 3. Configuration via URL

```go
func configurationExample() {
    // Passing configuration via URL parameters
    config := map[string]interface{}{
        "app": map[string]interface{}{
            "name":        "MyApp",
            "version":     "1.0.0",
            "environment": "production",
            "features": map[string]interface{}{
                "auth":       true,
                "analytics":  true,
                "dark_theme": false,
            },
        },
        "database": map[string]interface{}{
            "host":     "localhost",
            "port":     5432,
            "name":     "myapp_db",
            "ssl_mode": "require",
            "pool": map[string]interface{}{
                "min_connections": 5,
                "max_connections": 20,
            },
        },
        "cache": map[string]interface{}{
            "type": "redis",
            "ttl":  3600,
            "servers": []interface{}{
                "redis1.example.com:6379",
                "redis2.example.com:6379",
            },
        },
    }

    configString, _ := qs.Stringify(config)
    fmt.Printf("Configuration: %s\n", configString)

    // Parse back
    parsedConfig, _ := qs.Parse(configString)

    // Extract specific values
    if app, ok := parsedConfig["app"].(map[string]interface{}); ok {
        if features, ok := app["features"].(map[string]interface{}); ok {
            authEnabled := features["auth"]
            fmt.Printf("Authentication enabled: %v\n", authEnabled)
        }
    }
}
```

## Error handling and edge cases

```go
func errorHandlingExample() {
    // Handle invalid data
    malformedQuery := "invalid[bracket=value&another]=test"

    result, err := qs.Parse(malformedQuery)
    if err != nil {
        fmt.Printf("Parsing error: %v\n", err)
        return
    }

    fmt.Printf("Result: %+v\n", result)

    // Handle empty values
    emptyValues := "empty=&null&space= &encoded=%20"
    result2, _ := qs.Parse(emptyValues)
    fmt.Printf("Empty values: %+v\n", result2)

    // Handle special characters
    specialChars := "message=Привет, мир!&symbols=@#$%^&*()&emoji=🚀🔥💻"
    result3, _ := qs.Parse(specialChars)
    fmt.Printf("Special characters: %+v\n", result3)
}
```

## Performance and benchmarks

```go
func performanceExample() {
    // Test performance on large data
    largeData := make(map[string]interface{})

    // Create large nested structure
    for i := 0; i < 100; i++ {
        userKey := fmt.Sprintf("users[%d]", i)
        userData := map[string]interface{}{
            "id":   i,
            "name": fmt.Sprintf("User%d", i),
            "profile": map[string]interface{}{
                "email": fmt.Sprintf("user%d@example.com", i),
                "settings": map[string]interface{}{
                    "theme":         "dark",
                    "notifications": true,
                    "privacy": map[string]interface{}{
                        "show_email": false,
                        "show_phone": true,
                    },
                },
            },
            "tags": []interface{}{
                fmt.Sprintf("tag%d", i),
                fmt.Sprintf("category%d", i%10),
            },
        }
        largeData[userKey] = userData
    }

    // Measure time to create query string
    start := time.Now()
    queryString, _ := qs.Stringify(largeData)
    stringifyDuration := time.Since(start)

    // Measure time to parse
    start = time.Now()
    parsedData, _ := qs.Parse(queryString)
    parseDuration := time.Since(start)

    fmt.Printf("Query string size: %d characters\n", len(queryString))
    fmt.Printf("Stringify time: %v\n", stringifyDuration)
    fmt.Printf("Parse time: %v\n", parseDuration)
    fmt.Printf("Number of users: %d\n", len(parsedData))
}
```

## Integration with HTTP servers

### With standard library net/http

```go
func httpIntegrationExample() {
    http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
        // Parse query parameters
        queryString := r.URL.RawQuery
        params, err := qs.Parse(queryString)
        if err != nil {
            http.Error(w, "Invalid query parameters", http.StatusBadRequest)
            return
        }

        // Extract search parameters
        var searchQuery string
        var filters map[string]interface{}

        if query, ok := params["q"].(string); ok {
            searchQuery = query
        }

        if filterData, ok := params["filters"].(map[string]interface{}); ok {
            filters = filterData
        }

        // Search logic...
        results := performSearch(searchQuery, filters)

        // Send results
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(results)
    })
}

func performSearch(query string, filters map[string]interface{}) map[string]interface{} {
    // Stub for search logic
    return map[string]interface{}{
        "query":   query,
        "filters": filters,
        "results": []interface{}{},
        "total":   0,
    }
}
```

### With Gin Framework

```go
func ginIntegrationExample() {
    r := gin.Default()

    r.GET("/api/products", func(c *gin.Context) {
        // Get all query parameters
        queryString := c.Request.URL.RawQuery
        params, err := qs.Parse(queryString)
        if err != nil {
            c.JSON(400, gin.H{"error": "Invalid query parameters"})
            return
        }

        // Extract parameters
        var page, limit int = 1, 20
        var filters map[string]interface{}
        var sort map[string]interface{}

        if pagination, ok := params["page"].(map[string]interface{}); ok {
            if p, ok := pagination["number"].(string); ok {
                if parsed, err := strconv.Atoi(p); err == nil {
                    page = parsed
                }
            }
            if l, ok := pagination["size"].(string); ok {
                if parsed, err := strconv.Atoi(l); err == nil {
                    limit = parsed
                }
            }
        }

        if filterData, ok := params["filters"].(map[string]interface{}); ok {
            filters = filterData
        }

        if sortData, ok := params["sort"].(map[string]interface{}); ok {
            sort = sortData
        }

        // Business logic...
        products := getProducts(page, limit, filters, sort)

        c.JSON(200, gin.H{
            "data":       products,
            "pagination": gin.H{"page": page, "limit": limit},
            "filters":    filters,
            "sort":       sort,
        })
    })
}
```

## Struct Parsing with Query Tags

The Go qs library now supports parsing query strings directly into structs using `query` tags:

### Basic Struct Parsing

```go
package main

import (
    "fmt"
    "github.com/zaytra/qs/qs"
)

type User struct {
    Name     string  `query:"name"`
    Age      int     `query:"age"`
    Email    string  `query:"email"`
    IsActive bool    `query:"active"`
    Score    float64 `query:"score"`
}

func main() {
    queryString := "name=John&age=30&email=john@example.com&active=true&score=95.5"

    var user User
    err := qs.ParseToStruct(queryString, &user)
    if err != nil {
        panic(err)
    }

    fmt.Printf("User: %+v\n", user)
    // Output: User: {Name:John Age:30 Email:john@example.com IsActive:true Score:95.5}
}
```

### Struct with Arrays

```go
type SearchFilter struct {
    Query    string   `query:"q"`
    Tags     []string `query:"tags"`
    Category string   `query:"category"`
    MinPrice int      `query:"min_price"`
    MaxPrice int      `query:"max_price"`
}

func main() {
    queryString := "q=golang&tags[]=programming&tags[]=web&category=tech&min_price=10&max_price=100"

    var filter SearchFilter
    err := qs.ParseToStruct(queryString, &filter)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Filter: %+v\n", filter)
    // Output: Filter: {Query:golang Tags:[programming web] Category:tech MinPrice:10 MaxPrice:100}
}
```

### Nested Structs

```go
type NestedStruct struct {
    User     User              `query:"user"`
    Settings map[string]string `query:"settings"`
    Enabled  bool              `query:"enabled"`
}

func main() {
    queryString := "user[name]=Alice&user[age]=25&user[email]=alice@test.com&user[active]=false&user[score]=88.0&settings[theme]=dark&settings[lang]=en&enabled=true"

    var nested NestedStruct
    err := qs.ParseToStruct(queryString, &nested)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Nested: %+v\n", nested)
    // Output: Nested: {User:{Name:Alice Age:25 Email:alice@test.com IsActive:false Score:88} Settings:map[lang:en theme:dark] Enabled:true}
}
```

### Converting Struct to Query String

```go
func main() {
    user := &User{
        Name:     "John",
        Age:      30,
        Email:    "john@example.com",
        IsActive: true,
        Score:    95.5,
    }

    queryString, err := qs.StructToQueryString(user)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Query String: %s\n", queryString)
    // Output: Query String: active=true&age=30&email=john%40example.com&name=John&score=95.5
}
```

### Working with Maps and Structs

```go
func main() {
    // Convert struct to map
    user := &User{Name: "Charlie", Age: 28, Email: "charlie@example.com"}

    userMap, err := qs.StructToMap(user)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Map: %+v\n", userMap)
    // Output: Map: map[active:false age:28 email:charlie@example.com name:Charlie score:0]

    // Convert map back to struct
    var newUser User
    err = qs.MapToStruct(userMap, &newUser)
    if err != nil {
        panic(err)
    }

    fmt.Printf("New User: %+v\n", newUser)
    // Output: New User: {Name:Charlie Age:28 Email:charlie@example.com IsActive:false Score:0}
}
```

### Query Tag Options

- Use `query:"field_name"` to specify the query parameter name
- Use `query:"-"` to skip a field
- If no tag is provided, the field name in lowercase is used
- Supports all basic Go types: string, int, float, bool, slices, maps, and nested structs

### Supported Types

- **Basic types**: string, int (all variants), uint (all variants), float32/64, bool
- **Slices**: []string, []int, etc.
- **Maps**: map[string]string, map[string]interface{}
- **Nested structs**: Automatically converted to/from nested objects
- **Pointers**: Automatically handled

This library provides powerful capabilities for working with complex nested data structures in query strings, which is especially useful for APIs, forms, and analytical systems.

## Strapi-like API requests

### Complex example of API request in Strapi style

```go
// Structures for Strapi-like API
type StrapiQuery struct {
    Filters    map[string]interface{} `query:"filters"`
    Sort       []string               `query:"sort"`
    Fields     []string               `query:"fields"`
    Populate   map[string]interface{} `query:"populate"`
    Pagination map[string]interface{} `query:"pagination"`
    Locale     string                 `query:"locale"`
    Meta       map[string]interface{} `query:"meta"`
}

func strapiExample() {
    // Create complex Strapi-like request
    strapiQuery := StrapiQuery{
        Filters: map[string]interface{}{
            "title": map[string]interface{}{
                "$contains": "golang",
            },
            "publishedAt": map[string]interface{}{
                "$gte": "2023-01-01T00:00:00.000Z",
                "$lte": "2023-12-31T23:59:59.999Z",
            },
            "author": map[string]interface{}{
                "name": map[string]interface{}{
                    "$eq": "John Doe",
                },
                "role": map[string]interface{}{
                    "$in": []interface{}{"admin", "editor"},
                },
            },
            "$or": []interface{}{
                map[string]interface{}{
                    "featured": map[string]interface{}{
                        "$eq": true,
                    },
                },
                map[string]interface{}{
                    "views": map[string]interface{}{
                        "$gte": 1000,
                    },
                },
            },
            "categories": map[string]interface{}{
                "slug": map[string]interface{}{
                    "$in": []interface{}{"programming", "tutorials", "guides"},
                },
            },
            "tags": map[string]interface{}{
                "name": map[string]interface{}{
                    "$containsi": "react",
                },
            },
        },
        Sort: []string{"publishedAt:desc", "title:asc", "author.name:asc"},
        Fields: []string{"title", "content", "publishedAt", "slug", "featured"},
        Populate: map[string]interface{}{
            "author": map[string]interface{}{
                "fields": []interface{}{"name", "email", "avatar"},
                "populate": map[string]interface{}{
                    "avatar": map[string]interface{}{
                        "fields": []interface{}{"url", "alternativeText"},
                    },
                },
            },
            "categories": map[string]interface{}{
                "fields": []interface{}{"name", "slug", "description"},
                "sort":   []interface{}{"name:asc"},
            },
            "tags": map[string]interface{}{
                "fields": []interface{}{"name", "slug"},
                "filters": map[string]interface{}{
                    "name": map[string]interface{}{
                        "$ne": "deprecated",
                    },
                },
            },
            "cover": map[string]interface{}{
                "fields": []interface{}{"url", "width", "height", "formats"},
            },
            "comments": map[string]interface{}{
                "sort": []interface{}{"createdAt:desc"},
                "filters": map[string]interface{}{
                    "blocked": map[string]interface{}{
                        "$eq": false,
                    },
                },
                "populate": map[string]interface{}{
                    "author": map[string]interface{}{
                        "fields": []interface{}{"username", "email"},
                    },
                },
                "pagination": map[string]interface{}{
                    "page":     1,
                    "pageSize": 10,
                },
            },
        },
        Pagination: map[string]interface{}{
            "page":      1,
            "pageSize":  25,
            "withCount": true,
        },
        Locale: "en",
        Meta: map[string]interface{}{
            "analytics": true,
            "cache":     map[string]interface{}{"ttl": 300},
            "debug":     false,
        },
    }

    // Marshal to query string
    queryString, err := qs.Marshal(strapiQuery)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Strapi Query String (%d chars):\n%s\n\n", len(queryString), queryString)

    // Unmarshal back for verification
    var parsedQuery StrapiQuery
    err = qs.Unmarshal(queryString, &parsedQuery)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Parsed back successfully: %t\n", parsedQuery.Locale == "en")
}
```

## Integration with popular Go frameworks

### 1. Gin Framework

```go
package main

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/zaytra/qs/qs"
)

// Structures for API
type ArticleFilters struct {
    Title      string                 `query:"title"`
    Author     string                 `query:"author"`
    Status     string                 `query:"status"`
    DateRange  map[string]interface{} `query:"dateRange"`
    Categories []string               `query:"categories"`
    Tags       []string               `query:"tags"`
}

type ArticleQuery struct {
    Filters    ArticleFilters         `query:"filters"`
    Sort       []string               `query:"sort"`
    Fields     []string               `query:"fields"`
    Populate   []string               `query:"populate"`
    Pagination map[string]interface{} `query:"pagination"`
}

type Article struct {
    ID          int      `json:"id"`
    Title       string   `json:"title"`
    Content     string   `json:"content"`
    Author      string   `json:"author"`
    Status      string   `json:"status"`
    Categories  []string `json:"categories"`
    Tags        []string `json:"tags"`
    PublishedAt string   `json:"publishedAt"`
}

func setupGinServer() {
    r := gin.Default()

    // Middleware for parsing query parameters
    r.Use(func(c *gin.Context) {
        // Log complex query strings
        if len(c.Request.URL.RawQuery) > 100 {
            gin.Logger()(c)
        }
        c.Next()
    })

    // GET /api/articles - with extended filtering
    r.GET("/api/articles", getArticles)

    // POST /api/articles/search - search with request body
    r.POST("/api/articles/search", searchArticles)

    // GET /api/strapi-style - full Strapi-like endpoint
    r.GET("/api/strapi-style", getStrapiStyleData)

    r.Run(":8080")
}

func getArticles(c *gin.Context) {
    // Parse query parameters to structure
    var query ArticleQuery
    if err := qs.Unmarshal(c.Request.URL.RawQuery, &query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Invalid query parameters",
            "details": err.Error(),
        })
        return
    }

    // Set default values
    if len(query.Sort) == 0 {
        query.Sort = []string{"publishedAt:desc"}
    }

    if query.Pagination == nil {
        query.Pagination = map[string]interface{}{
            "page":     1,
            "pageSize": 20,
        }
    }

    // Simulate data retrieval from DB
    articles := getArticlesFromDB(query)
    total := getTotalArticlesCount(query.Filters)

    // Create meta-information for response
    meta := map[string]interface{}{
        "pagination": map[string]interface{}{
            "page":      query.Pagination["page"],
            "pageSize":  query.Pagination["pageSize"],
            "total":     total,
            "pageCount": (total + getPageSize(query.Pagination) - 1) / getPageSize(query.Pagination),
        },
        "filters": query.Filters,
        "sort":    query.Sort,
    }

    // Create query string for next page
    nextPageQuery := query
    if page, ok := query.Pagination["page"].(int); ok {
        nextPageQuery.Pagination["page"] = page + 1
    }
    nextPageURL, _ := qs.Marshal(nextPageQuery)

    c.JSON(http.StatusOK, gin.H{
        "data": articles,
        "meta": meta,
        "links": gin.H{
            "self": c.Request.URL.String(),
            "next": "/api/articles?" + nextPageURL,
        },
    })
}

func searchArticles(c *gin.Context) {
    // Parse JSON from request body
    var searchRequest map[string]interface{}
    if err := c.ShouldBindJSON(&searchRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Convert to query string for logging/caching
    searchQuery, _ := qs.Marshal(searchRequest)

    // Parse back to typed structure
    var query ArticleQuery
    if err := qs.Unmarshal(searchQuery, &query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    articles := getArticlesFromDB(query)

    c.JSON(http.StatusOK, gin.H{
        "data":         articles,
        "searchQuery":  searchQuery,
        "originalBody": searchRequest,
    })
}

func getStrapiStyleData(c *gin.Context) {
    // Full Strapi-like endpoint
    var strapiQuery StrapiQuery
    if err := qs.Unmarshal(c.Request.URL.RawQuery, &strapiQuery); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Simulate complex Strapi-request processing
    data := processStrapiQuery(strapiQuery)

    // Create Strapi-style response
    response := gin.H{
        "data": data,
        "meta": gin.H{
            "pagination": strapiQuery.Pagination,
            "filters":    strapiQuery.Filters,
            "sort":       strapiQuery.Sort,
            "populate":   strapiQuery.Populate,
        },
    }

    c.JSON(http.StatusOK, response)
}

// Helper functions (stubs)
func getArticlesFromDB(query ArticleQuery) []Article {
    return []Article{
        {
            ID:          1,
            Title:       "Getting Started with Go",
            Content:     "Introduction to Go programming...",
            Author:      "John Doe",
            Status:      "published",
            Categories:  []string{"programming", "golang"},
            Tags:        []string{"beginner", "tutorial"},
            PublishedAt: "2023-11-15T10:30:00Z",
        },
        {
            ID:          2,
            Title:       "Advanced Go Patterns",
            Content:     "Deep dive into Go patterns...",
            Author:      "Jane Smith",
            Status:      "published",
            Categories:  []string{"programming", "golang", "advanced"},
            Tags:        []string{"patterns", "advanced"},
            PublishedAt: "2023-11-10T14:20:00Z",
        },
    }
}

func getTotalArticlesCount(filters ArticleFilters) int {
    return 42 // stub
}

func getPageSize(pagination map[string]interface{}) int {
    if pageSize, ok := pagination["pageSize"].(int); ok {
        return pageSize
    }
    return 20
}

func processStrapiQuery(query StrapiQuery) []map[string]interface{} {
    return []map[string]interface{}{
        {
            "id":    1,
            "title": "Sample Article",
            "author": map[string]interface{}{
                "name":  "John Doe",
                "email": "john@example.com",
            },
        },
    }
}
```

### 2. Echo Framework

```go
package main

import (
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/zaytra/qs/qs"
)

func setupEchoServer() {
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Custom middleware for parsing query
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Add parsed query to context
            if rawQuery := c.Request().URL.RawQuery; rawQuery != "" {
                var parsedQuery map[string]interface{}
                if err := qs.Unmarshal(rawQuery, &parsedQuery); err == nil {
                    c.Set("parsedQuery", parsedQuery)
                }
            }
            return next(c)
        }
    })

    // Routes
    e.GET("/api/products", getProducts)
    e.GET("/api/complex-search", complexSearch)
    e.POST("/api/query-builder", queryBuilder)

    e.Logger.Fatal(e.Start(":8081"))
}

type ProductQuery struct {
    Filters struct {
        Name        string                 `query:"name"`
        Category    string                 `query:"category"`
        PriceRange  map[string]interface{} `query:"priceRange"`
        InStock     bool                   `query:"inStock"`
        Brand       []string               `query:"brand"`
        Rating      map[string]interface{} `query:"rating"`
        Attributes  map[string]interface{} `query:"attributes"`
        Availability map[string]interface{} `query:"availability"`
    } `query:"filters"`
    Sort       []string               `query:"sort"`
    Fields     []string               `query:"fields"`
    Include    []string               `query:"include"`
    Pagination map[string]interface{} `query:"pagination"`
    Facets     map[string]interface{} `query:"facets"`
}

func getProducts(c echo.Context) error {
    // Parse to typed structure
    var query ProductQuery
    if err := qs.Unmarshal(c.Request().URL.RawQuery, &query); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
            "error":   "Invalid query parameters",
            "details": err.Error(),
            "query":   c.Request().URL.RawQuery,
        })
    }

    // Validate and set default values
    if err := validateProductQuery(&query); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
            "error": "Validation failed",
            "details": err.Error(),
        })
    }

    // Get data
    products := getProductsFromDB(query)
    facets := getFacetsForQuery(query)
    total := getTotalProductsCount(query)

    // Build response
    response := map[string]interface{}{
        "data":    products,
        "facets":  facets,
        "total":   total,
        "query":   query,
        "timing": map[string]interface{}{
            "total": "45ms",
            "db":    "32ms",
            "cache": "13ms",
        },
    }

    return c.JSON(http.StatusOK, response)
}

func complexSearch(c echo.Context) error {
    // Get already parsed query from middleware
    parsedQuery, ok := c.Get("parsedQuery").(map[string]interface{})
    if !ok {
        return echo.NewHTTPError(http.StatusBadRequest, "No query parameters provided")
    }

    // Complex search logic with multiple conditions
    searchResults := performComplexSearch(parsedQuery)

    // Create query string for caching
    cacheKey, _ := qs.Marshal(parsedQuery)

    response := map[string]interface{}{
        "results":     searchResults,
        "originalQuery": parsedQuery,
        "cacheKey":    cacheKey,
        "suggestions": generateSearchSuggestions(parsedQuery),
    }

    return c.JSON(http.StatusOK, response)
}

func queryBuilder(c echo.Context) error {
    // POST endpoint for building complex query
    var requestBody map[string]interface{}
    if err := c.Bind(&requestBody); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    // Convert to query string
    builtQuery, err := qs.Marshal(requestBody)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    // Create full URL
    baseURL := c.Scheme() + "://" + c.Request().Host + "/api/products"
    fullURL := baseURL + "?" + builtQuery

    // Parse back for validation
    var validatedQuery ProductQuery
    if err := qs.Unmarshal(builtQuery, &validatedQuery); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
            "error": "Generated query is invalid",
            "query": builtQuery,
        })
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "originalRequest": requestBody,
        "generatedQuery":  builtQuery,
        "fullURL":         fullURL,
        "validatedQuery":  validatedQuery,
        "queryLength":     len(builtQuery),
    })
}

// Helper functions
func validateProductQuery(query *ProductQuery) error {
    // Set default values
    if query.Pagination == nil {
        query.Pagination = map[string]interface{}{
            "page":     1,
            "pageSize": 20,
        }
    }

    if len(query.Sort) == 0 {
        query.Sort = []string{"relevance:desc", "price:asc"}
    }

    // Validate pagination
    if page, ok := query.Pagination["page"].(float64); ok && page < 1 {
        query.Pagination["page"] = 1
    }

    if pageSize, ok := query.Pagination["pageSize"].(float64); ok && (pageSize < 1 || pageSize > 100) {
        query.Pagination["pageSize"] = 20
    }

    return nil
}

func getProductsFromDB(query ProductQuery) []map[string]interface{} {
    // Stub for getting products
    return []map[string]interface{}{
        {
            "id":       1,
            "name":     "MacBook Pro",
            "category": "electronics",
            "price":    2499.99,
            "inStock":  true,
            "brand":    "Apple",
            "rating":   4.8,
        },
        {
            "id":       2,
            "name":     "iPhone 15",
            "category": "electronics",
            "price":    999.99,
            "inStock":  true,
            "brand":    "Apple",
            "rating":   4.9,
        },
    }
}

func getFacetsForQuery(query ProductQuery) map[string]interface{} {
    return map[string]interface{}{
        "brands": map[string]int{
            "Apple":   15,
            "Samsung": 12,
            "Google":  8,
        },
        "categories": map[string]int{
            "electronics": 35,
            "computers":   20,
            "phones":      15,
        },
        "priceRanges": map[string]int{
            "0-500":    10,
            "500-1000": 15,
            "1000+":    20,
        },
    }
}

func getTotalProductsCount(query ProductQuery) int {
    return 156 // stub
}

func performComplexSearch(query map[string]interface{}) []map[string]interface{} {
    return []map[string]interface{}{
        {"id": 1, "score": 0.95, "type": "product"},
        {"id": 2, "score": 0.87, "type": "product"},
    }
}

func generateSearchSuggestions(query map[string]interface{}) []string {
    return []string{
        "Try adding more filters",
        "Consider different price range",
        "Check similar categories",
    }
}
```

### 3. Example usage

```go
func main() {
    // Demonstrate creating complex Strapi-like request
    fmt.Println("=== Strapi-style Query Example ===")
    strapiExample()

    // Start servers (choose one)
    fmt.Println("\nStarting Gin server on :8080...")
    go setupGinServer()

    fmt.Println("Starting Echo server on :8081...")
    go setupEchoServer()

    // Wait
    select {}
}
```

### Example queries to servers

```bash
# Simple query with filters
curl "http://localhost:8080/api/articles?filters[author]=John&filters[status]=published&sort[]=publishedAt:desc&pagination[page]=1&pagination[pageSize]=10"

# Complex Strapi-like query
curl "http://localhost:8080/api/strapi-style?filters[title][\$contains]=golang&filters[author][role][\$in][]=admin&filters[author][role][\$in][]=editor&populate[author][fields][]=name&populate[author][fields][]=email&sort[]=publishedAt:desc&pagination[page]=1&pagination[pageSize]=25"

# Query to Echo server
curl "http://localhost:8081/api/products?filters[name]=MacBook&filters[priceRange][min]=1000&filters[priceRange][max]=3000&filters[brand][]=Apple&filters[brand][]=Dell&sort[]=price:asc&pagination[page]=1"

# POST query for building query
curl -X POST "http://localhost:8081/api/query-builder" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": {
      "category": "electronics",
      "priceRange": {"min": 500, "max": 2000},
      "inStock": true
    },
    "sort": ["price:asc", "rating:desc"],
    "pagination": {"page": 1, "pageSize": 20}
  }'
```

These examples show the full power of the QS library when working with modern Go web frameworks and complex API structures in Strapi style!

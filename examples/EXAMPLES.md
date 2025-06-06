# QS Go Library - Usage Examples

The QS library for Go provides powerful capabilities for working with query strings, including parsing and creating complex nested data structures.

## Table of Contents

1. [Strapi-style API Examples](#strapi-style-api-examples)
   - [Quick Start](#quick-start)
   - [Complex Strapi-like Structures](#complex-strapi-like-structures)
   - [HTTP Handler Integration](#http-handler-integration)
   - [Typical Usage Scenarios](#typical-strapi-usage-scenarios)
   - [Performance](#strapi-performance)
   - [Complex Strapi Example](#complex-strapi-example)
2. [Idiomatic Marshal/Unmarshal Functions](#idiomatic-marshalunmarshal-functions)
3. [Basic Examples](#basic-examples)
4. [Complex Nested Structures](#complex-nested-structures)
5. [Working with Options](#working-with-options)
6. [Real-world Scenarios](#real-world-scenarios)
7. [Struct Parsing with Query Tags](#struct-parsing-with-query-tags)
8. [Error Handling and Edge Cases](#error-handling-and-edge-cases)
9. [Performance and Benchmarks](#performance-and-benchmarks)
10. [HTTP Server Integration](#integration-with-http-servers)

## Strapi-style API Examples

This section demonstrates using the QS library to create Strapi-like APIs with rich query string structures.

### Quick Start

```bash
# Run main example with Strapi demo
go run example_struct.go strapi_demo.go

# Or just Strapi demo (uncomment main function first)
go run strapi_demo.go
```

### Complex Strapi-like Structures

```go
type StrapiQuery struct {
    Filters    map[string]interface{} `query:"filters"`
    Sort       []string               `query:"sort"`
    Fields     []string               `query:"fields"`
    Populate   map[string]interface{} `query:"populate"`
    Pagination map[string]interface{} `query:"pagination"`
    Locale     string                 `query:"locale"`
    Meta       map[string]interface{} `query:"meta"`
}
```

### Automatic Type Detection with Strapi Queries

```go
// Single API for all data types
var strapiQuery StrapiQuery
qs.Unmarshal(queryString, &strapiQuery)  // struct

var mapData map[string]interface{}
qs.Unmarshal(queryString, &mapData)      // map

queryString, _ := qs.Marshal(anyData)    // any type
```

### HTTP Handler Integration

```go
func getArticles(c *gin.Context) {
    var query ArticleQuery
    if err := qs.Unmarshal(c.Request.URL.RawQuery, &query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    articles := getArticlesFromDB(query)
    c.JSON(200, gin.H{"data": articles, "meta": createMeta(query)})
}
```

### Example Generated Strapi Query String

```
filters[$or][0][featured][$eq]=true&filters[$or][1][views][$gte]=1000&filters[author][name][$eq]=John%20Doe&filters[author][role][$in][0]=admin&filters[author][role][$in][1]=editor&filters[categories][slug][$in][0]=programming&filters[categories][slug][$in][1]=tutorials&filters[categories][slug][$in][2]=guides&filters[publishedAt][$gte]=2023-01-01T00:00:00.000Z&filters[publishedAt][$lte]=2023-12-31T23:59:59.999Z&filters[title][$contains]=golang&fields[0]=title&fields[1]=content&fields[2]=publishedAt&fields[3]=slug&fields[4]=featured&locale=en&meta[analytics]=true&meta[cache][ttl]=300&meta[debug]=false&pagination[page]=1&pagination[pageSize]=25&pagination[withCount]=true&populate[author][fields][0]=name&populate[author][fields][1]=email&populate[author][fields][2]=avatar&populate[author][populate][avatar][fields][0]=url&populate[author][populate][avatar][fields][1]=alternativeText&populate[categories][fields][0]=name&populate[categories][fields][1]=slug&populate[categories][fields][2]=description&populate[categories][sort][0]=name:asc&populate[comments][filters][blocked][$eq]=false&populate[comments][pagination][page]=1&populate[comments][pagination][pageSize]=10&populate[comments][sort][0]=createdAt:desc&sort[0]=publishedAt:desc&sort[1]=title:asc&sort[2]=author.name:asc
```

### Supported Frameworks

- **Gin** - the most popular Go web framework
- **Echo** - high-performance minimalist framework
- **Chi** - lightweight router
- **net/http** - standard library

### Typical Strapi Usage Scenarios

#### E-commerce API
```bash
/api/products?filters[category]=electronics&filters[price][min]=100&filters[price][max]=1000&sort[]=price:asc&pagination[page]=1
```

#### Content Management
```bash
/api/articles?filters[status]=published&filters[author][role][$in][]=admin&populate[author][fields][]=name&sort[]=publishedAt:desc
```

#### Analytics Dashboard
```bash
/api/analytics?filters[dateRange][start]=2023-01-01&filters[dateRange][end]=2023-12-31&fields[]=pageviews&fields[]=sessions
```

### Strapi Performance

Benchmark results for complex Strapi-like structures:

```
BenchmarkMarshalComplex-10    406468    2964 ns/op
BenchmarkUnmarshalComplex-10   75964   15018 ns/op
```

### Strapi Features

- ✅ **Runtime type detection** - automatic type detection
- ✅ **Deep nesting support** - deep structure nesting
- ✅ **Strapi compatibility** - compatibility with Strapi API
- ✅ **Framework agnostic** - works with any framework
- ✅ **High performance** - optimized performance
- ✅ **Rich query structures** - support for complex queries

### Comparison with JavaScript qs

This Go library is fully compatible with the popular JavaScript [qs](https://github.com/ljharb/qs) library, but adds Go's typing and performance.

```javascript
// JavaScript qs
const qs = require('qs');
const query = qs.stringify({filters: {status: 'published'}});

// Go qs with typing
var query ArticleQuery
qs.Unmarshal(queryString, &query)  // Type safety!
```

### Complex Strapi Example

Here's a comprehensive example showing a complex Strapi-like query with all features:

```go
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

## Idiomatic Marshal/Unmarshal Functions

### Automatic type detection at runtime

The new `Marshal` and `Unmarshal` functions automatically detect data types at runtime and choose the appropriate processing method.

```go
package main

import (
    "fmt"
    "github.com/zaytracom/qs/qs"
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
    "github.com/zaytracom/qs/qs"
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

## Struct Parsing with Query Tags

The Go qs library now supports parsing query strings directly into structs using `query` tags:

### Basic Struct Parsing

```go
package main

import (
    "fmt"
    "github.com/zaytracom/qs/qs"
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

**Note:** For comprehensive framework integration examples including Gin, Echo, and other popular Go frameworks, see the detailed examples in the main [Strapi-style API Examples](#strapi-style-api-examples) section above.

## Summary

This comprehensive examples file demonstrates the full power of the QS library for Go, covering:

- **Strapi-style APIs** with complex nested structures and operators
- **Idiomatic Marshal/Unmarshal** with automatic type detection
- **Framework integration** with Gin, Echo, Chi, and net/http
- **Real-world scenarios** from e-commerce to analytics
- **Performance optimization** and benchmarking
- **Error handling** and edge cases
- **Struct parsing** with query tags

The QS library provides a powerful, type-safe, and performant solution for working with complex query strings in Go applications, making it easy to build sophisticated APIs and data processing systems.

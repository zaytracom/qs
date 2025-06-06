# QS Library - Strapi-style API Demo

This example demonstrates using the QS library to create Strapi-like APIs with rich query string structures.

## Quick Start

```bash
# Run main example with Strapi demo
go run example_struct.go strapi_demo.go

# Or just Strapi demo
go run strapi_demo.go
```

## What this example demonstrates

### 1. Complex Strapi-like structures

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

### 2. Automatic type detection

```go
// Single API for all data types
var strapiQuery StrapiQuery
qs.Unmarshal(queryString, &strapiQuery)  // struct

var mapData map[string]interface{}
qs.Unmarshal(queryString, &mapData)      // map

queryString, _ := qs.Marshal(anyData)    // any type
```

### 3. HTTP Handler integration

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

## Example generated query string

```
filters[$or][0][featured][$eq]=true&filters[$or][1][views][$gte]=1000&filters[author][name][$eq]=John%20Doe&filters[author][role][$in][0]=admin&filters[author][role][$in][1]=editor&filters[categories][slug][$in][0]=programming&filters[categories][slug][$in][1]=tutorials&filters[categories][slug][$in][2]=guides&filters[publishedAt][$gte]=2023-01-01T00:00:00.000Z&filters[publishedAt][$lte]=2023-12-31T23:59:59.999Z&filters[title][$contains]=golang&fields[0]=title&fields[1]=content&fields[2]=publishedAt&fields[3]=slug&fields[4]=featured&locale=en&meta[analytics]=true&meta[cache][ttl]=300&meta[debug]=false&pagination[page]=1&pagination[pageSize]=25&pagination[withCount]=true&populate[author][fields][0]=name&populate[author][fields][1]=email&populate[author][fields][2]=avatar&populate[author][populate][avatar][fields][0]=url&populate[author][populate][avatar][fields][1]=alternativeText&populate[categories][fields][0]=name&populate[categories][fields][1]=slug&populate[categories][fields][2]=description&populate[categories][sort][0]=name:asc&populate[comments][filters][blocked][$eq]=false&populate[comments][pagination][page]=1&populate[comments][pagination][pageSize]=10&populate[comments][sort][0]=createdAt:desc&sort[0]=publishedAt:desc&sort[1]=title:asc&sort[2]=author.name:asc
```

## Supported frameworks

- **Gin** - the most popular Go web framework
- **Echo** - high-performance minimalist framework
- **Chi** - lightweight router
- **net/http** - standard library

## Typical usage scenarios

### E-commerce API
```bash
/api/products?filters[category]=electronics&filters[price][min]=100&filters[price][max]=1000&sort[]=price:asc&pagination[page]=1
```

### Content Management
```bash
/api/articles?filters[status]=published&filters[author][role][$in][]=admin&populate[author][fields][]=name&sort[]=publishedAt:desc
```

### Analytics Dashboard
```bash
/api/analytics?filters[dateRange][start]=2023-01-01&filters[dateRange][end]=2023-12-31&fields[]=pageviews&fields[]=sessions
```

## Performance

Benchmark results for complex structures:

```
BenchmarkMarshalComplex-10    406468    2964 ns/op
BenchmarkUnmarshalComplex-10   75964   15018 ns/op
```

## Features

- ✅ **Runtime type detection** - automatic type detection
- ✅ **Deep nesting support** - deep structure nesting
- ✅ **Strapi compatibility** - compatibility with Strapi API
- ✅ **Framework agnostic** - works with any framework
- ✅ **High performance** - optimized performance
- ✅ **Rich query structures** - support for complex queries

## Comparison with JavaScript qs

This Go library is fully compatible with the popular JavaScript [qs](https://github.com/ljharb/qs) library, but adds Go's typing and performance.

```javascript
// JavaScript qs
const qs = require('qs');
const query = qs.stringify({filters: {status: 'published'}});

// Go qs with typing
var query ArticleQuery
qs.Unmarshal(queryString, &query)  // Type safety!
```

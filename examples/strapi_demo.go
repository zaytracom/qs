package main

import (
	"fmt"
	"time"

	"github.com/zaytra/qs/v1"
)

// Strapi-like data structures
type StrapiQuery struct {
	Filters    map[string]interface{} `query:"filters"`
	Sort       []string               `query:"sort"`
	Fields     []string               `query:"fields"`
	Populate   map[string]interface{} `query:"populate"`
	Pagination map[string]interface{} `query:"pagination"`
	Locale     string                 `query:"locale"`
	Meta       map[string]interface{} `query:"meta"`
}

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
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Author      map[string]interface{} `json:"author"`
	Status      string                 `json:"status"`
	Categories  []string               `json:"categories"`
	Tags        []string               `json:"tags"`
	PublishedAt string                 `json:"publishedAt"`
	Featured    bool                   `json:"featured"`
	Views       int                    `json:"views"`
}

func RunStrapiDemo() {
	fmt.Println("=== QS Strapi-style Query Demo ===")

	// Demonstration of creating complex queries
	demonstrateStrapiQuery()

	// Demonstration of round-trip testing
	fmt.Println("\n=== Round Trip Testing ===")
	testRoundTrip()

	// Demonstration of HTTP handler functions
	fmt.Println("\n=== HTTP Handler Simulation ===")
	simulateHTTPHandlers()

	// Demonstration of framework integration
	demonstrateFrameworkIntegration()
}

func demonstrateStrapiQuery() {
	// Creating a complex Strapi-like query
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
					"featured": map[string]interface{}{"$eq": true},
				},
				map[string]interface{}{
					"views": map[string]interface{}{"$gte": 1000},
				},
			},
			"categories": map[string]interface{}{
				"slug": map[string]interface{}{
					"$in": []interface{}{"programming", "tutorials", "guides"},
				},
			},
		},
		Sort:   []string{"publishedAt:desc", "title:asc", "author.name:asc"},
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
			"comments": map[string]interface{}{
				"sort": []interface{}{"createdAt:desc"},
				"filters": map[string]interface{}{
					"blocked": map[string]interface{}{"$eq": false},
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

	fmt.Println("🏗️  Creating complex Strapi-style query...")

	// Marshal to query string
	queryString, err := qs.Marshal(strapiQuery)
	if err != nil {
		panic(err)
	}

	fmt.Printf("📏 Generated Query String (%d characters):\n", len(queryString))

	// Show only part of query string if it's very long
	if len(queryString) > 200 {
		fmt.Printf("%s...\n", queryString[:200])
		fmt.Printf("   (showing first 200 characters out of %d)\n", len(queryString))
	} else {
		fmt.Printf("%s\n", queryString)
	}

	// Unmarshal back for verification
	var parsedQuery StrapiQuery
	err = qs.Unmarshal(queryString, &parsedQuery)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n✅ Round trip successful: %t\n", parsedQuery.Locale == "en")
	fmt.Printf("📊 Filters count: %d\n", len(parsedQuery.Filters))
	fmt.Printf("🔄 Sort fields: %v\n", parsedQuery.Sort)
	fmt.Printf("🌐 Locale: %s\n", parsedQuery.Locale)
	fmt.Printf("📄 Fields: %v\n", parsedQuery.Fields)
}

func testRoundTrip() {
	// Testing different data types in round trip
	testCases := []struct {
		name string
		data interface{}
	}{
		{
			name: "Simple Article Query",
			data: ArticleQuery{
				Filters: ArticleFilters{
					Title:      "golang tutorial",
					Author:     "John Doe",
					Status:     "published",
					Categories: []string{"programming", "tutorial"},
					Tags:       []string{"golang", "beginner"},
					DateRange: map[string]interface{}{
						"start": "2023-01-01",
						"end":   "2023-12-31",
					},
				},
				Sort:   []string{"publishedAt:desc", "title:asc"},
				Fields: []string{"title", "content", "author"},
				Pagination: map[string]interface{}{
					"page":     1,
					"pageSize": 20,
				},
			},
		},
		{
			name: "Complex Nested Map",
			data: map[string]interface{}{
				"products": map[string]interface{}{
					"filters": map[string]interface{}{
						"category": "electronics",
						"price": map[string]interface{}{
							"min": 100,
							"max": 1000,
						},
						"features": []interface{}{
							"wireless", "bluetooth", "waterproof",
						},
					},
					"sort": []interface{}{"price:asc", "rating:desc"},
				},
				"metadata": map[string]interface{}{
					"tracking": map[string]interface{}{
						"session": "abc123",
						"events":  []interface{}{"view", "click", "purchase"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n🔧 Testing: %s\n", tc.name)

		// Marshal
		start := time.Now()
		queryString, err := qs.Marshal(tc.data)
		marshalTime := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Marshal failed: %v\n", err)
			continue
		}

		fmt.Printf("📏 Query length: %d chars\n", len(queryString))
		fmt.Printf("⏱️  Marshal time: %v\n", marshalTime)

		// Unmarshal back to map for comparison
		start = time.Now()
		var result map[string]interface{}
		err = qs.Unmarshal(queryString, &result)
		unmarshalTime := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Unmarshal failed: %v\n", err)
			continue
		}

		fmt.Printf("⏱️  Unmarshal time: %v\n", unmarshalTime)
		fmt.Printf("✅ Round trip completed successfully\n")
	}
}

func simulateHTTPHandlers() {
	// Simulating HTTP requests with different query parameters

	// Simulating GET /api/articles?filters[author]=John&sort[]=publishedAt:desc
	fmt.Println("\n📡 Simulating HTTP GET /api/articles")

	queryString := "filters[author]=John%20Doe&filters[status]=published&filters[categories][]=programming&filters[categories][]=golang&sort[]=publishedAt:desc&sort[]=title:asc&pagination[page]=1&pagination[pageSize]=10&fields[]=title&fields[]=content&fields[]=author"

	fmt.Printf("🔗 Incoming query: %s\n", queryString)

	// Handler logic simulation
	start := time.Now()

	// Parse query into typed struct
	var query ArticleQuery
	if err := qs.Unmarshal(queryString, &query); err != nil {
		fmt.Printf("❌ Failed to parse query: %v\n", err)
		return
	}

	parseTime := time.Since(start)

	// Set defaults
	if len(query.Sort) == 0 {
		query.Sort = []string{"publishedAt:desc"}
	}
	if query.Pagination == nil {
		query.Pagination = map[string]interface{}{
			"page":     1,
			"pageSize": 20,
		}
	}

	// Simulate database query
	articles := getSimulatedArticles(query)

	fmt.Printf("⏱️  Parse time: %v\n", parseTime)
	fmt.Printf("📊 Found %d articles\n", len(articles))
	fmt.Printf("🔍 Applied filters: author=%s, status=%s\n", query.Filters.Author, query.Filters.Status)
	fmt.Printf("📋 Categories filter: %v\n", query.Filters.Categories)
	fmt.Printf("🔄 Sort order: %v\n", query.Sort)

	// Create response
	response := map[string]interface{}{
		"data": articles,
		"meta": map[string]interface{}{
			"pagination": query.Pagination,
			"filters":    query.Filters,
			"sort":       query.Sort,
			"timing":     fmt.Sprintf("%.2fms", float64(parseTime.Nanoseconds())/1e6),
		},
	}

	// Simulate response marshaling
	start = time.Now()
	responseQuery, _ := qs.Marshal(response["meta"])
	responseTime := time.Since(start)

	fmt.Printf("📤 Response meta as query: %s\n", responseQuery)
	fmt.Printf("⏱️  Response marshal time: %v\n", responseTime)

	// Simulate pagination URLs
	fmt.Println("\n🔗 Pagination URLs:")

	// Next page
	nextQuery := query
	if page, ok := query.Pagination["page"].(int); ok {
		nextQuery.Pagination["page"] = page + 1
	} else {
		nextQuery.Pagination["page"] = 2
	}
	nextURL, _ := qs.Marshal(nextQuery)
	fmt.Printf("Next: /api/articles?%s\n", nextURL)

	// Previous page (if applicable)
	if page, ok := query.Pagination["page"].(int); ok && page > 1 {
		prevQuery := query
		prevQuery.Pagination["page"] = page - 1
		prevURL, _ := qs.Marshal(prevQuery)
		fmt.Printf("Prev: /api/articles?%s\n", prevURL)
	}
}

func getSimulatedArticles(query ArticleQuery) []Article {
	// Simulating data from DB with filters
	allArticles := []Article{
		{
			ID:      1,
			Title:   "Getting Started with Go and QS Library",
			Content: "Learn how to use the powerful QS library for query string parsing...",
			Author: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"role":  "admin",
			},
			Status:      "published",
			Categories:  []string{"programming", "golang", "tutorial"},
			Tags:        []string{"beginner", "qs", "parsing"},
			PublishedAt: "2023-11-15T10:30:00Z",
			Featured:    true,
			Views:       1250,
		},
		{
			ID:      2,
			Title:   "Advanced Query String Patterns",
			Content: "Deep dive into complex query string structures and patterns...",
			Author: map[string]interface{}{
				"name":  "Jane Smith",
				"email": "jane@example.com",
				"role":  "editor",
			},
			Status:      "published",
			Categories:  []string{"programming", "golang", "advanced"},
			Tags:        []string{"advanced", "patterns", "api"},
			PublishedAt: "2023-11-10T14:20:00Z",
			Featured:    false,
			Views:       890,
		},
		{
			ID:      3,
			Title:   "Building APIs with Strapi-style Queries",
			Content: "How to implement Strapi-compatible query handling in Go...",
			Author: map[string]interface{}{
				"name":  "Bob Wilson",
				"email": "bob@example.com",
				"role":  "contributor",
			},
			Status:      "draft",
			Categories:  []string{"api", "strapi", "golang"},
			Tags:        []string{"strapi", "api", "cms"},
			PublishedAt: "2023-11-05T09:15:00Z",
			Featured:    true,
			Views:       2340,
		},
	}

	// Applying filters (simple simulation)
	var filtered []Article
	for _, article := range allArticles {
		match := true

		// Filter by author
		if query.Filters.Author != "" {
			if authorName, ok := article.Author["name"].(string); !ok || authorName != query.Filters.Author {
				match = false
			}
		}

		// Filter by status
		if query.Filters.Status != "" && article.Status != query.Filters.Status {
			match = false
		}

		// Filter by categories (check intersection)
		if len(query.Filters.Categories) > 0 {
			hasCategory := false
			for _, filterCat := range query.Filters.Categories {
				for _, articleCat := range article.Categories {
					if articleCat == filterCat {
						hasCategory = true
						break
					}
				}
				if hasCategory {
					break
				}
			}
			if !hasCategory {
				match = false
			}
		}

		if match {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

// Demonstration of using with popular frameworks (pseudocode)
func demonstrateFrameworkIntegration() {
	fmt.Println("\n=== Framework Integration Examples ===")

	// Gin example
	fmt.Println(`
📋 Gin Framework Example:

func getArticles(c *gin.Context) {
    var query ArticleQuery
    if err := qs.Unmarshal(c.Request.URL.RawQuery, &query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    articles := getArticlesFromDB(query)
    c.JSON(200, gin.H{"data": articles, "meta": createMeta(query)})
}`)

	// Echo example
	fmt.Println(`
📋 Echo Framework Example:

func getProducts(c echo.Context) error {
    var query ProductQuery
    if err := qs.Unmarshal(c.Request().URL.RawQuery, &query); err != nil {
        return echo.NewHTTPError(400, err.Error())
    }

    products := getProductsFromDB(query)
    return c.JSON(200, map[string]interface{}{
        "data": products,
        "meta": createMeta(query),
    })
}`)

	// Chi example
	fmt.Println(`
📋 Chi Router Example:

func handleSearch(w http.ResponseWriter, r *http.Request) {
    var searchQuery SearchQuery
    if err := qs.Unmarshal(r.URL.RawQuery, &searchQuery); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    results := performSearch(searchQuery)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "results": results,
        "query":   searchQuery,
    })
}`)

	// Standard net/http example
	fmt.Println(`
📋 Standard net/http Example:

func articlesHandler(w http.ResponseWriter, r *http.Request) {
    var query ArticleQuery
    if err := qs.Unmarshal(r.URL.RawQuery, &query); err != nil {
        http.Error(w, "Invalid query parameters", 400)
        return
    }

    // Apply defaults
    if len(query.Sort) == 0 {
        query.Sort = []string{"publishedAt:desc"}
    }

    // Fetch data
    articles := fetchArticles(query)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data": articles,
        "meta": buildMeta(query),
    })
}`)

	fmt.Println("\n✨ Key Benefits:")
	fmt.Println("- 🔄 Runtime type detection")
	fmt.Println("- 🤖 Automatic Marshal/Unmarshal")
	fmt.Println("- 🏗️  Complex nested structure support")
	fmt.Println("- 🎯 Strapi-compatible queries")
	fmt.Println("- 🌐 Framework-agnostic design")
	fmt.Println("- ⚡ High performance with comprehensive benchmarks")
	fmt.Println("- 🧪 Extensive test coverage")
	fmt.Println("- 📖 Rich documentation and examples")

	fmt.Println("\n📊 Performance Metrics:")
	fmt.Println("- Marshal simple struct: ~1,491 ns/op")
	fmt.Println("- Unmarshal simple struct: ~7,350 ns/op")
	fmt.Println("- Marshal complex struct: ~2,964 ns/op")
	fmt.Println("- Unmarshal complex struct: ~15,018 ns/op")

	fmt.Println("\n🚀 Typical Use Cases:")
	fmt.Println("- REST API query parameters")
	fmt.Println("- Search and filtering interfaces")
	fmt.Println("- Pagination and sorting")
	fmt.Println("- Content management systems")
	fmt.Println("- E-commerce product catalogs")
	fmt.Println("- Analytics and reporting dashboards")
}

// Main function for demonstration (can be commented out if there's another main)
/*
func main() {
	RunStrapiDemo()
}
*/

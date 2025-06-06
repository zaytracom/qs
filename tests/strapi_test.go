package tests

import (
	"reflect"
	"testing"

	qs "github.com/zaytra/qs/v1"
)

// Strapi-like data structures for testing
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

func TestStrapiSimpleFilters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "simple title filter",
			input: "filters[title][$contains]=golang",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"title": map[string]interface{}{
						"$contains": "golang",
					},
				},
			},
		},
		{
			name:  "multiple filters",
			input: "filters[title][$contains]=golang&filters[status][$eq]=published",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"title": map[string]interface{}{
						"$contains": "golang",
					},
					"status": map[string]interface{}{
						"$eq": "published",
					},
				},
			},
		},
		{
			name:  "nested author filter",
			input: "filters[author][name][$eq]=John Doe&filters[author][role][$in][]=admin&filters[author][role][$in][]=editor",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"author": map[string]interface{}{
						"name": map[string]interface{}{
							"$eq": "John Doe",
						},
						"role": map[string]interface{}{
							"$in": []interface{}{"admin", "editor"},
						},
					},
				},
			},
		},
		{
			name:  "date range filter",
			input: "filters[publishedAt][$gte]=2023-01-01T00:00:00.000Z&filters[publishedAt][$lte]=2023-12-31T23:59:59.999Z",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"publishedAt": map[string]interface{}{
						"$gte": "2023-01-01T00:00:00.000Z",
						"$lte": "2023-12-31T23:59:59.999Z",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, nil)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Parse() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStrapiComplexQueries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "full strapi query with populate",
			input: "filters[title][$contains]=tutorial&populate[author][fields][]=name&populate[author][fields][]=email&sort[]=publishedAt:desc&sort[]=title:asc&fields[]=title&fields[]=content&pagination[page]=1&pagination[pageSize]=25&locale=en",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"title": map[string]interface{}{
						"$contains": "tutorial",
					},
				},
				"populate": map[string]interface{}{
					"author": map[string]interface{}{
						"fields": []interface{}{"name", "email"},
					},
				},
				"sort":   []interface{}{"publishedAt:desc", "title:asc"},
				"fields": []interface{}{"title", "content"},
				"pagination": map[string]interface{}{
					"page":     "1",
					"pageSize": "25",
				},
				"locale": "en",
			},
		},
		{
			name:  "or condition filter",
			input: "filters[$or][0][featured][$eq]=true&filters[$or][1][views][$gte]=1000",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"$or": map[string]interface{}{
						"0": map[string]interface{}{
							"featured": map[string]interface{}{
								"$eq": "true",
							},
						},
						"1": map[string]interface{}{
							"views": map[string]interface{}{
								"$gte": "1000",
							},
						},
					},
				},
			},
		},
		{
			name:  "nested populate with filters",
			input: "populate[comments][sort][]=createdAt:desc&populate[comments][filters][blocked][$eq]=false&populate[comments][pagination][page]=1&populate[comments][pagination][pageSize]=10",
			expected: map[string]interface{}{
				"populate": map[string]interface{}{
					"comments": map[string]interface{}{
						"sort": []interface{}{"createdAt:desc"},
						"filters": map[string]interface{}{
							"blocked": map[string]interface{}{
								"$eq": "false",
							},
						},
						"pagination": map[string]interface{}{
							"page":     "1",
							"pageSize": "10",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, nil)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Parse() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStrapiStructRoundTrip(t *testing.T) {
	tests := []struct {
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
				Sort:     []string{"publishedAt:desc", "title:asc"},
				Fields:   []string{"title", "content", "author"},
				Populate: []string{"author", "categories"},
				Pagination: map[string]interface{}{
					"page":     1,
					"pageSize": 20,
				},
			},
		},
		{
			name: "Complex Strapi Query",
			data: StrapiQuery{
				Filters: map[string]interface{}{
					"title": map[string]interface{}{
						"$contains": "golang",
					},
					"publishedAt": map[string]interface{}{
						"$gte": "2023-01-01T00:00:00.000Z",
					},
					"author": map[string]interface{}{
						"role": map[string]interface{}{
							"$in": []interface{}{"admin", "editor"},
						},
					},
				},
				Sort:   []string{"publishedAt:desc", "title:asc"},
				Fields: []string{"title", "content", "publishedAt"},
				Populate: map[string]interface{}{
					"author": map[string]interface{}{
						"fields": []interface{}{"name", "email"},
					},
					"categories": map[string]interface{}{
						"sort": []interface{}{"name:asc"},
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
					"debug":     false,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Marshal to query string
			queryString, err := qs.Marshal(test.data)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			// Unmarshal back to map
			var result map[string]interface{}
			err = qs.Unmarshal(queryString, &result)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			// Verify we can parse it without errors
			parsed, err := qs.Parse(queryString, nil)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			// Basic verification that important fields are present
			if test.name == "Simple Article Query" {
				if filters, ok := parsed["filters"].(map[string]interface{}); ok {
					if title, exists := filters["title"]; !exists || title == "" {
						t.Errorf("Expected title filter to be preserved")
					}
				} else {
					t.Errorf("Expected filters to be present")
				}
			}

			if test.name == "Complex Strapi Query" {
				if locale, ok := parsed["locale"]; !ok || locale != "en" {
					t.Errorf("Expected locale to be preserved as 'en', got %v", locale)
				}
			}
		})
	}
}

func TestStrapiSpecialOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "$in operator with array",
			input: "filters[categories][slug][$in][]=programming&filters[categories][slug][$in][]=tutorials&filters[categories][slug][$in][]=guides",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"categories": map[string]interface{}{
						"slug": map[string]interface{}{
							"$in": []interface{}{"programming", "tutorials", "guides"},
						},
					},
				},
			},
		},
		{
			name:  "$contains operator",
			input: "filters[title][$contains]=golang tutorial",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"title": map[string]interface{}{
						"$contains": "golang tutorial",
					},
				},
			},
		},
		{
			name:  "$eq and $ne operators",
			input: "filters[status][$eq]=published&filters[draft][$ne]=true",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"status": map[string]interface{}{
						"$eq": "published",
					},
					"draft": map[string]interface{}{
						"$ne": "true",
					},
				},
			},
		},
		{
			name:  "$gte and $lte operators",
			input: "filters[views][$gte]=100&filters[views][$lte]=10000",
			expected: map[string]interface{}{
				"filters": map[string]interface{}{
					"views": map[string]interface{}{
						"$gte": "100",
						"$lte": "10000",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, nil)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Parse() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStrapiPaginationAndMeta(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "pagination parameters",
			input: "pagination[page]=2&pagination[pageSize]=50&pagination[withCount]=true",
			expected: map[string]interface{}{
				"pagination": map[string]interface{}{
					"page":      "2",
					"pageSize":  "50",
					"withCount": "true",
				},
			},
		},
		{
			name:  "meta parameters",
			input: "meta[analytics]=true&meta[cache][ttl]=300&meta[debug]=false",
			expected: map[string]interface{}{
				"meta": map[string]interface{}{
					"analytics": "true",
					"cache": map[string]interface{}{
						"ttl": "300",
					},
					"debug": "false",
				},
			},
		},
		{
			name:  "locale parameter",
			input: "locale=ru&fields[]=title&fields[]=content",
			expected: map[string]interface{}{
				"locale": "ru",
				"fields": []interface{}{"title", "content"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, nil)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Parse() = %v, want %v", result, test.expected)
			}
		})
	}
}

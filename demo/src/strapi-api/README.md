# Strapi API Query Example

This example demonstrates how to build Strapi-compatible query strings using the `qs` library.

Strapi uses a specific query string format for filtering, sorting, pagination, and population of related data. The key option here is `encodeValuesOnly: true` (JS) / `WithEncodeValuesOnly(true)` (Go), which keeps bracket characters in keys unencoded for better readability while still encoding special characters in values.

## Filtering with operators

JS:

```js
qs.stringify(
  {
    filters: {
      title: { $contains: "hello" },
      createdAt: { $gte: "2023-01-01" },
    },
  },
  { encodeValuesOnly: true }
)
// filters[title][$contains]=hello&filters[createdAt][$gte]=2023-01-01
```

Go:

```go
qs.Stringify(
  map[string]any{
    "filters": map[string]any{
      "title":     map[string]any{"$contains": "hello"},
      "createdAt": map[string]any{"$gte": "2023-01-01"},
    },
  },
  qs.WithEncodeValuesOnly(true),
)
// filters[title][$contains]=hello&filters[createdAt][$gte]=2023-01-01
```

## Sorting

JS:

```js
qs.stringify({ sort: ["title:asc", "createdAt:desc"] }, { encodeValuesOnly: true })
// sort[0]=title:asc&sort[1]=createdAt:desc
```

Go:

```go
qs.Stringify(
  map[string]any{"sort": []any{"title:asc", "createdAt:desc"}},
  qs.WithEncodeValuesOnly(true),
)
// sort[0]=title:asc&sort[1]=createdAt:desc
```

## Pagination

JS:

```js
qs.stringify(
  { pagination: { page: 1, pageSize: 10 } },
  { encodeValuesOnly: true }
)
// pagination[page]=1&pagination[pageSize]=10
```

Go:

```go
qs.Stringify(
  map[string]any{"pagination": map[string]any{"page": 1, "pageSize": 10}},
  qs.WithEncodeValuesOnly(true),
)
// pagination[page]=1&pagination[pageSize]=10
```

## Population (nested relations)

JS:

```js
qs.stringify(
  {
    populate: {
      author: { fields: ["name", "email"] },
      categories: { fields: ["name"] },
    },
  },
  { encodeValuesOnly: true }
)
// populate[author][fields][0]=name&populate[author][fields][1]=email&populate[categories][fields][0]=name
```

Go:

```go
qs.Stringify(
  map[string]any{
    "populate": map[string]any{
      "author":     map[string]any{"fields": []any{"name", "email"}},
      "categories": map[string]any{"fields": []any{"name"}},
    },
  },
  qs.WithEncodeValuesOnly(true),
)
// populate[author][fields][0]=name&populate[author][fields][1]=email&populate[categories][fields][0]=name
```

## Complete Strapi query

JS:

```js
qs.stringify(
  {
    filters: { status: { $eq: "published" } },
    sort: ["createdAt:desc"],
    pagination: { page: 1, pageSize: 25 },
    populate: { author: { fields: ["name"] } },
  },
  { encodeValuesOnly: true, sort: (a, b) => a.localeCompare(b) }
)
// filters[status][$eq]=published&pagination[page]=1&pagination[pageSize]=25&populate[author][fields][0]=name&sort[0]=createdAt:desc
```

Go:

```go
qs.Stringify(
  map[string]any{
    "filters":    map[string]any{"status": map[string]any{"$eq": "published"}},
    "sort":       []any{"createdAt:desc"},
    "pagination": map[string]any{"page": 1, "pageSize": 25},
    "populate":   map[string]any{"author": map[string]any{"fields": []any{"name"}}},
  },
  qs.WithEncodeValuesOnly(true),
  qs.WithSort(func(a, b string) bool { return a < b }),
)
// filters[status][$eq]=published&pagination[page]=1&pagination[pageSize]=25&populate[author][fields][0]=name&sort[0]=createdAt:desc
```


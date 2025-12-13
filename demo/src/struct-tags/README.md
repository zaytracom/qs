# Go structs + `query` tags (Marshal/Unmarshal)

In Go, you can work not only with `map[string]any`, but also with typed structs using `query` struct tags.

Use `qs.Marshal` to convert a struct to a query string, and `qs.Unmarshal` to parse a query string directly into a struct. Field names are specified via `` `query:"name"` `` tags. Both functions support all the same options as `Stringify` and `Parse` respectively.

## Marshal struct → query string

Go:

```go
type Filters struct {
  Status string   `query:"status" json:"status"`
  Tags   []string `query:"tags"   json:"tags"`
}

type Query struct {
  Filters Filters `query:"filters" json:"filters"`
  Page    int     `query:"page"    json:"page"`
}

q := Query{
  Filters: Filters{Status: "published", Tags: []string{"go", "qs"}},
  Page:    2,
}

qs.Marshal(
  q,
  qs.WithStringifyAllowDots(true),
  qs.WithStringifyArrayFormat(qs.ArrayFormatBrackets),
  qs.WithStringifyEncode(false),
  qs.WithStringifySort(func(a, b string) bool { return a < b }),
)
// filters.status=published&filters.tags[]=go&filters.tags[]=qs&page=2
```

JS (equivalent form):

```js
qs.stringify(
  { filters: { status: "published", tags: ["go", "qs"] }, page: 2 },
  {
    allowDots: true,
    arrayFormat: "brackets",
    encode: false,
    sort: (a, b) => a.localeCompare(b),
  }
)
// filters.status=published&filters.tags[]=go&filters.tags[]=qs&page=2
```

## Unmarshal query string → struct

Go:

```go
var out Query
qs.Unmarshal(
  "filters.status=published&filters.tags[]=go&filters.tags[]=qs&page=2",
  &out,
  qs.WithParseAllowDots(true),
)
// {"filters":{"status":"published","tags":["go","qs"]},"page":2}
```

JS:

```js
const parsed = qs.parse("filters.status=published&filters.tags[]=go&filters.tags[]=qs&page=2", {
  allowDots: true,
})
parsed.page = Number(parsed.page)
console.log(JSON.stringify(parsed))
// {"filters":{"status":"published","tags":["go","qs"]},"page":2}
```

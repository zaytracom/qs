# sort / Sort

The `sort` option (JS) / `WithStringifySort(...)` (Go, Stringify) controls key ordering during stringification.

## Stringify

JS:

```js
qs.stringify({ b: 2, a: 1 }, { sort: (a, b) => a.localeCompare(b), encode: false })
// a=1&b=2
```

Go:

```go
qs.Stringify(
  map[string]any{"b": 2, "a": 1},
  qs.WithStringifySort(func(a, b string) bool { return a < b }),
  qs.WithStringifyEncode(false),
)
// a=1&b=2
```

## Parse

Sort is stringify-only.


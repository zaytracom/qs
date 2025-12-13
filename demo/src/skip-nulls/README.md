# skipNulls / SkipNulls

The `skipNulls` option (JS) / `WithStringifySkipNulls(true)` (Go, Stringify) removes keys with null/nil values from the output.

## Stringify

JS:

```js
qs.stringify({ a: null, b: "x" }, { skipNulls: true, encode: false })
// b=x
```

Go:

```go
qs.Stringify(
  map[string]any{"a": nil, "b": "x"},
  qs.WithStringifySkipNulls(true),
  qs.WithStringifyEncode(false),
)
// b=x
```

## Parse

This is stringify-only.


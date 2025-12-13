# addQueryPrefix / AddQueryPrefix

The `addQueryPrefix` option (JS) / `WithStringifyAddQueryPrefix(true)` (Go) adds a leading `?` when stringifying.

By default, `qs.stringify` returns just the key-value pairs (e.g., `a=b&c=d`). When building a complete URL, you often need the `?` prefix. Instead of manually concatenating `"?" + qs.stringify(...)`, enable this option to include it automatically.

JS:

```js
qs.stringify({ a: "b" }, { encode: false })
// a=b

qs.stringify({ a: "b" }, { addQueryPrefix: true, encode: false })
// ?a=b
```

Go:

```go
qs.Stringify(map[string]any{"a": "b"}, qs.WithStringifyEncode(false))
// a=b

qs.Stringify(
  map[string]any{"a": "b"},
  qs.WithStringifyAddQueryPrefix(true),
  qs.WithStringifyEncode(false),
)
// ?a=b
```

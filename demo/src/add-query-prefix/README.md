# addQueryPrefix / AddQueryPrefix

The `addQueryPrefix` option (JS) / `WithStringifyAddQueryPrefix(true)` (Go) adds a leading `?` when stringifying.

JS:

```js
qs.stringify({ a: "b" }, { encode: false })
// a=b

qs.stringify({ a: "b" }, { addQueryPrefix: true, encode: false })
// ?a=b
```

Go:

```go
qs.Stringify(map[string]any{"a": "b"}, qs.WithEncode(false))
// a=b

qs.Stringify(
  map[string]any{"a": "b"},
  qs.WithStringifyAddQueryPrefix(true),
  qs.WithEncode(false),
)
// ?a=b
```


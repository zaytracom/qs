# encode / Encode

The `encode` option (JS) / `WithStringifyEncode(false)` (Go, Stringify) enables/disables percent-encoding during stringification.

## Stringify

JS:

```js
qs.stringify({ a: "a b" })
// a=a%20b

qs.stringify({ a: "a b" }, { encode: false })
// a=a b
```

Go:

```go
qs.Stringify(map[string]any{"a": "a b"})
// a=a%20b

qs.Stringify(map[string]any{"a": "a b"}, qs.WithStringifyEncode(false))
// a=a b
```

## Parse

Parse always decodes percent-encoding; there is no `encode` option for parsing.


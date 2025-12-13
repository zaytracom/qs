# encodeValuesOnly / EncodeValuesOnly

The `encodeValuesOnly` option (JS) / `WithStringifyEncodeValuesOnly(true)` (Go, Stringify) encodes values but keeps keys readable (e.g., brackets stay as `[`/`]`).

## Stringify

JS:

```js
qs.stringify({ a: { b: "x y" } }, { encodeValuesOnly: true })
// a[b]=x%20y
```

Go:

```go
qs.Stringify(
  map[string]any{"a": map[string]any{"b": "x y"}},
  qs.WithStringifyEncodeValuesOnly(true),
)
// a[b]=x%20y
```

## Parse

This is stringify-only. Parsing always accepts both encoded and unencoded bracket characters.


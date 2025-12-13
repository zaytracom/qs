# encoder / Encoder

The `encoder` option (JS) / `WithStringifyEncoder(...)` (Go, Stringify) lets you override how strings are encoded during stringification.

## Stringify

JS:

```js
qs.stringify({ a: "b" }, { encoder: (str) => str.toUpperCase(), encode: true })
// A=B
```

Go:

```go
qs.Stringify(
  map[string]any{"a": "b"},
  qs.WithStringifyEncoder(func(s string, charset qs.Charset, kind string, format qs.Format) string {
    return strings.ToUpper(s)
  }),
)
// A=B
```

## Parse

Parsing uses `decoder`: see `../decoder/README.md`.


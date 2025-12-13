# decoder / Decoder

The `decoder` option (JS) / `WithParseDecoder(...)` (Go, Parse) lets you override how percent-encoded strings are decoded.

## Parse

JS:

```js
qs.parse("a=b", {
  decoder: (str, defaultDecoder, charset, type) => str.toUpperCase(),
})
// {"a":"B"}
```

Go:

```go
qs.Parse(
  "a=b",
  qs.WithParseDecoder(func(s string, charset qs.Charset, kind string) (string, error) {
    return strings.ToUpper(s), nil
  }),
)
// {"a":"B"}
```

## Stringify

Stringification uses `encoder` instead: see `../encoder/README.md`.


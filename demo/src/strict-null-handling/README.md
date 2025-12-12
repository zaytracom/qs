# strictNullHandling / StrictNullHandling

The `strictNullHandling` option (JS) / `WithStrictNullHandling(true)` (Go, Parse) and `WithStringifyStrictNullHandling(true)` (Go, Stringify) changes how `null` is handled:

- when stringifying, `null` becomes a key without `=` (e.g., `a` instead of `a=`)
- when parsing, a key without a value is treated as `null`, not as an empty string

## Stringify (without and with strictNullHandling)

JS:

```js
qs.stringify({ a: null }, { encode: false })
// a=

qs.stringify({ a: null }, { strictNullHandling: true, encode: false })
// a
```

Go:

```go
qs.Stringify(map[string]any{"a": nil}, qs.WithEncode(false))
// a=

qs.Stringify(map[string]any{"a": nil}, qs.WithStringifyStrictNullHandling(true), qs.WithEncode(false))
// a
```

## Parse (without and with strictNullHandling)

JS:

```js
qs.parse("a&b=")
// {"a":"","b":""}

qs.parse("a&b=", { strictNullHandling: true })
// {"a":null,"b":""}
```

Go:

```go
qs.Parse("a&b=")
// {"a":"","b":""}

qs.Parse("a&b=", qs.WithStrictNullHandling(true))
// {"a":null,"b":""}
```


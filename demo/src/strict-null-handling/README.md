# strictNullHandling / StrictNullHandling

The `strictNullHandling` option (JS) / `WithStrictNullHandling(true)` (Go, Parse) and `WithStringifyStrictNullHandling(true)` (Go, Stringify) changes how `null` is handled.

By default, `null` values are stringified as empty strings (`a=`) and keys without values are parsed as empty strings. With `strictNullHandling`, you can distinguish between `null` and empty string: `null` is serialized without `=` (just `a`), and keys without `=` are parsed back as `null`.

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


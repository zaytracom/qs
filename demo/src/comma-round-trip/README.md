# commaRoundTrip / CommaRoundTrip

The `commaRoundTrip` option (JS) / `WithStringifyCommaRoundTrip(true)` (Go, Stringify) affects how **single-element arrays** are serialized when using comma array format.

## Stringify

JS:

```js
qs.stringify({ a: ["x"] }, { arrayFormat: "comma", encode: false })
// a=x

qs.stringify({ a: ["x"] }, { arrayFormat: "comma", commaRoundTrip: true, encode: false })
// a[]=x
```

Go:

```go
qs.Stringify(
  map[string]any{"a": []any{"x"}},
  qs.WithStringifyArrayFormat(qs.ArrayFormatComma),
  qs.WithStringifyEncode(false),
)
// a=x

qs.Stringify(
  map[string]any{"a": []any{"x"}},
  qs.WithStringifyArrayFormat(qs.ArrayFormatComma),
  qs.WithStringifyCommaRoundTrip(true),
  qs.WithStringifyEncode(false),
)
// a[]=x
```

## Parse

This option is stringify-only.


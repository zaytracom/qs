# charsetSentinel / CharsetSentinel

The `charsetSentinel` option (JS) / `WithParseCharsetSentinel(true)` (Go, Parse) and `WithStringifyCharsetSentinel(true)` (Go, Stringify) adds/reads a `utf8=✓`-style parameter that indicates the intended charset.

## Stringify

JS:

```js
qs.stringify({ a: "1" }, { charsetSentinel: true, encode: false })
// utf8=✓&a=1
```

Go:

```go
qs.Stringify(
  map[string]any{"a": "1"},
  qs.WithStringifyCharsetSentinel(true),
  qs.WithStringifyEncode(false),
)
// utf8=✓&a=1
```

## Parse

JS (ISO-8859-1 sentinel):

```js
qs.parse("utf8=%26%2310003%3B&a=%E9", { charsetSentinel: true })
// {"a":"é"}
```

Go:

```go
qs.Parse(
  "utf8=%26%2310003%3B&a=%E9",
  qs.WithParseCharsetSentinel(true),
)
// {"a":"é"}
```


# interpretNumericEntities / InterpretNumericEntities

The `interpretNumericEntities` option (JS) / `WithParseInterpretNumericEntities(true)` (Go, Parse) turns HTML numeric entities like `&#9786;` into characters when using ISO-8859-1 decoding.

## Parse

JS:

```js
qs.parse("a=%26%239786%3B", { charset: "iso-8859-1" })
// {"a":"&#9786;"}

qs.parse("a=%26%239786%3B", { charset: "iso-8859-1", interpretNumericEntities: true })
// {"a":"☺"}
```

Go:

```go
qs.Parse("a=%26%239786%3B", qs.WithParseCharset(qs.CharsetISO88591))
// {"a":"&#9786;"}

qs.Parse(
  "a=%26%239786%3B",
  qs.WithParseCharset(qs.CharsetISO88591),
  qs.WithParseInterpretNumericEntities(true),
)
// {"a":"☺"}
```

## Stringify

This is a parse-time option; stringification is unaffected.


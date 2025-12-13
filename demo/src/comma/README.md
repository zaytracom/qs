# comma / Comma

The `comma` option (JS) / `WithParseComma(true)` (Go, Parse) enables parsing comma-separated values into arrays.

## Parse

JS:

```js
qs.parse("a=1,2")
// {"a":"1,2"}

qs.parse("a=1,2", { comma: true })
// {"a":["1","2"]}
```

Go:

```go
qs.Parse("a=1,2")
// {"a":"1,2"}

qs.Parse("a=1,2", qs.WithParseComma(true))
// {"a":["1","2"]}
```

## Stringify

In Go, comma output is controlled by `ArrayFormatComma`:

```go
qs.Stringify(
  map[string]any{"a": []any{"1", "2"}},
  qs.WithStringifyArrayFormat(qs.ArrayFormatComma),
  qs.WithStringifyEncode(false),
)
// a=1,2
```


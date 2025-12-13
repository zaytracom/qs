# decodeDotInKeys / DecodeDotInKeys

The `decodeDotInKeys` option (JS) / `WithParseDecodeDotInKeys(true)` (Go, Parse) treats `%2E` in keys as a dot token, enabling dot-notation parsing for encoded dots.

## Parse

JS:

```js
qs.parse("a%2Eb=c", { allowDots: true, decodeDotInKeys: true })
// {"a":{"b":"c"}}
```

Go:

```go
qs.Parse("a%2Eb=c", qs.WithParseDecodeDotInKeys(true))
// {"a":{"b":"c"}}
```

## Stringify

For the output side, see `encodeDotInKeys`:

```go
qs.Stringify(
  map[string]any{"a.b": "c"},
  qs.WithStringifyEncodeDotInKeys(true),
)
// a%2Eb=c
```


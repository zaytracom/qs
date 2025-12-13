# encodeDotInKeys / EncodeDotInKeys

The `encodeDotInKeys` option (JS) / `WithStringifyEncodeDotInKeys(true)` (Go, Stringify) percent-encodes `.` in **literal key names** as `%2E`.

This helps avoid ambiguity when dot-notation is used elsewhere.

## Stringify

JS:

```js
qs.stringify({ "a.b": "c" }, { encodeDotInKeys: true, encode: false })
// a%2Eb=c
```

Go:

```go
qs.Stringify(
  map[string]any{"a.b": "c"},
  qs.WithStringifyEncodeDotInKeys(true),
  qs.WithStringifyEncode(false),
)
// a%2Eb=c
```

## Parse

On the parse side, use `decodeDotInKeys`: see `../decode-dot-in-keys/README.md`.


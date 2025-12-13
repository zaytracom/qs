# strictDepth / StrictDepth

The `strictDepth` option (JS) / `WithParseStrictDepth(true)` (Go, Parse) turns depth overflow into an error instead of treating the remainder as a literal key.

## Parse

JS:

```js
qs.parse("a[b][c]=d", { depth: 1, strictDepth: true })
// throws (depth limit exceeded)
```

Go:

```go
qs.Parse(
  "a[b][c]=d",
  qs.WithParseDepth(1),
  qs.WithParseStrictDepth(true),
)
// error: depth limit exceeded
```

## Stringify

This is a parse-time option; stringification is unaffected.


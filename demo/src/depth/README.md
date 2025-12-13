# depth / Depth

The `depth` option (JS) / `WithParseDepth(n)` (Go, Parse) limits nesting depth when parsing keys with bracket/dot notation.

When the depth limit is hit, the remainder is treated as literal key text (unless `strictDepth` is enabled).

## Parse

JS:

```js
qs.parse("a[b][c][d]=x", { depth: 2 })
// {"a":{"b":{"c":{"[d]":"x"}}}}
```

Go:

```go
qs.Parse("a[b][c][d]=x", qs.WithParseDepth(2))
// {"a":{"b":{"c":{"[d]":"x"}}}}
```

## Stringify

Depth is a parse-time option; stringification is unaffected.


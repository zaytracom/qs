# arrayLimit / ArrayLimit

The `arrayLimit` option (JS) / `WithParseArrayLimit(n)` (Go, Parse) limits which numeric bracket segments are treated as array indices.

If an index is larger than `arrayLimit`, it is treated as a string key (object) instead of an array index.

## Parse

JS:

```js
qs.parse("a[2]=x", { arrayLimit: 1 })
// {"a":{"2":"x"}}
```

Go:

```go
qs.Parse("a[2]=x", qs.WithParseArrayLimit(1))
// {"a":{"2":"x"}}
```

## Stringify

This is a parse-time option; stringification is unaffected (use `ArrayFormat` for output shape).


# allowSparse / AllowSparse

The `allowSparse` option (JS) / `WithParseAllowSparse(true)` (Go, Parse) controls whether sparse array indices are preserved.

## Parse

JS:

```js
qs.parse("a[1]=b&a[3]=c")
// {"a":["b","c"]}

qs.parse("a[1]=b&a[3]=c", { allowSparse: true })
// {"a":[null,"b",null,"c"]}
```

Go:

```go
qs.Parse("a[1]=b&a[3]=c")
// {"a":["b","c"]}

qs.Parse("a[1]=b&a[3]=c", qs.WithParseAllowSparse(true))
// {"a":[nil,"b",nil,"c"]}
```

## Stringify

This is a parse-time semantic option; stringification is unaffected.


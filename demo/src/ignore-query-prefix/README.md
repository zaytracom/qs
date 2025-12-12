# ignoreQueryPrefix / IgnoreQueryPrefix

The `ignoreQueryPrefix` option (JS) / `WithIgnoreQueryPrefix(true)` (Go) ignores the leading `?` when parsing a query string.

JS:

```js
qs.parse("?a=b&c=d")
// {"?a":"b","c":"d"}

qs.parse("?a=b&c=d", { ignoreQueryPrefix: true })
// {"a":"b","c":"d"}
```

Go:

```go
qs.Parse("?a=b&c=d")
// {"?a":"b","c":"d"}

qs.Parse("?a=b&c=d", qs.WithIgnoreQueryPrefix(true))
// {"a":"b","c":"d"}
```


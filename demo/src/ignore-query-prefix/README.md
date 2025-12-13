# ignoreQueryPrefix / IgnoreQueryPrefix

The `ignoreQueryPrefix` option (JS) / `WithParseIgnoreQueryPrefix(true)` (Go) ignores the leading `?` when parsing a query string.

When parsing `window.location.search` or similar URL parts, the string often starts with `?`. Without this option, the `?` becomes part of the first key name. Enable `ignoreQueryPrefix` to automatically strip it.

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

qs.Parse("?a=b&c=d", qs.WithParseIgnoreQueryPrefix(true))
// {"a":"b","c":"d"}
```

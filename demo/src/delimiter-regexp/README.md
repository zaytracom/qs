# delimiter / DelimiterRegexp

JS `qs` supports regexp delimiters; Go `qs/v2` exposes this as `WithParseDelimiterRegexp(...)` (Parse-only).

## Parse

JS:

```js
qs.parse("a=1;b=2&c=3", { delimiter: /[;&]/ })
// {"a":"1","b":"2","c":"3"}
```

Go:

```go
re := regexp.MustCompile(`[;&]`)
qs.Parse("a=1;b=2&c=3", qs.WithParseDelimiterRegexp(re))
// {"a":"1","b":"2","c":"3"}
```

## Stringify

Stringify uses a string delimiter only: see `../delimiter/README.md`.


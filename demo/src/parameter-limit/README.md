# parameterLimit / ParameterLimit

The `parameterLimit` option (JS) / `WithParseParameterLimit(n)` (Go, Parse) limits how many parameters are processed.

## Parse

JS:

```js
qs.parse("a=1&b=2&c=3", { parameterLimit: 2 })
// {"a":"1","b":"2"}
```

Go:

```go
qs.Parse("a=1&b=2&c=3", qs.WithParseParameterLimit(2))
// {"a":"1","b":"2"}
```

To turn this into an error, also enable `throwOnLimitExceeded`:

```go
qs.Parse(
  "a=1&b=2&c=3",
  qs.WithParseParameterLimit(2),
  qs.WithParseThrowOnLimitExceeded(true),
)
// error: parameter limit exceeded
```

## Stringify

This is a parse-time option; stringification is unaffected.


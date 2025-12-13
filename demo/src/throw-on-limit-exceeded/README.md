# throwOnLimitExceeded / ThrowOnLimitExceeded

The `throwOnLimitExceeded` option (JS) / `WithParseThrowOnLimitExceeded(true)` (Go, Parse) makes limit violations return errors instead of being ignored.

## Parse (ParameterLimit)

JS:

```js
qs.parse("a=1&b=2&c=3", { parameterLimit: 2, throwOnLimitExceeded: true })
// throws (parameter limit exceeded)
```

Go:

```go
qs.Parse(
  "a=1&b=2&c=3",
  qs.WithParseParameterLimit(2),
  qs.WithParseThrowOnLimitExceeded(true),
)
// error: parameter limit exceeded
```

## Parse (ArrayLimit via duplicate combine)

Go:

```go
qs.Parse(
  "a=1&a=2&a=3",
  qs.WithParseArrayLimit(2),
  qs.WithParseThrowOnLimitExceeded(true),
)
// error: array limit exceeded
```

## Stringify

This is a parse-time option; stringification is unaffected.


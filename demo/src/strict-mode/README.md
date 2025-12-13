# strictMode / StrictMode

The `strictMode` option (Go-only) / `WithParseStrictMode(true)` enables strict syntax validation while parsing.

It reports errors for malformed keys/values (unclosed/unmatched brackets, invalid percent-encoding, dot-notation edge cases, etc.).

## Parse

Go:

```go
_, err := qs.Parse("a[b=1", qs.WithParseStrictMode(true))
// err == "unclosed bracket"
```

## Stringify

Strict mode is parse-only.


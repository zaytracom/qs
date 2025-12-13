# format / Format

The `format` option (JS) / `WithStringifyFormat(...)` (Go, Stringify) selects RFC 1738 vs RFC 3986 encoding rules.

## Stringify

JS:

```js
qs.stringify({ a: "a b" }, { format: "RFC3986" })
// a=a%20b

qs.stringify({ a: "a b" }, { format: "RFC1738" })
// a=a+b
```

Go:

```go
qs.Stringify(map[string]any{"a": "a b"}, qs.WithStringifyFormat(qs.FormatRFC3986))
// a=a%20b

qs.Stringify(map[string]any{"a": "a b"}, qs.WithStringifyFormat(qs.FormatRFC1738))
// a=a+b
```

## Parse

Parsing accepts both `%20` and `+` (where applicable); this option affects output formatting.


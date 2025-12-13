# serializeDate / SerializeDate

The `serializeDate` option (JS) / `WithStringifySerializeDate(...)` (Go, Stringify) controls how `Date` / `time.Time` values are turned into strings.

## Stringify

JS:

```js
qs.stringify({ t: new Date("2025-01-02T03:04:05Z") })
// t=2025-01-02T03%3A04%3A05.000Z (implementation-specific)
```

Go:

```go
t := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
qs.Stringify(
  map[string]any{"t": t},
  qs.WithStringifySerializeDate(func(t time.Time) string {
    return t.UTC().Format(time.RFC3339)
  }),
)
// t=2025-01-02T03%3A04%3A05Z
```

## Parse

This is stringify-only.


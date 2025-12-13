# filter / Filter

The `filter` option (JS) / `WithStringifyFilter(...)` (Go, Stringify) can be:

- a function that transforms/skips values, or
- a list of allowed keys.

## Stringify (allowlist)

JS:

```js
qs.stringify({ a: 1, b: 2 }, { filter: ["a"], encode: false })
// a=1
```

Go:

```go
qs.Stringify(
  map[string]any{"a": 1, "b": 2},
  qs.WithStringifyFilter([]string{"a"}),
  qs.WithStringifyEncode(false),
)
// a=1
```

## Stringify (function)

JS:

```js
qs.stringify({ a: 1, b: 2 }, { filter: (prefix, value) => (prefix === "b" ? undefined : value), encode: false })
// a=1
```

Go:

```go
qs.Stringify(
  map[string]any{"a": 1, "b": 2},
  qs.WithStringifyFilter(func(prefix string, value any) any {
    if prefix == "b" {
      return nil
    }
    return value
  }),
  qs.WithStringifyEncode(false),
)
// a=1
```

## Parse

Filter is stringify-only.


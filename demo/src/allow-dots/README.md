# allowDots / AllowDots

The `allowDots` option (JS) / `WithAllowDots(true)` (Go, Parse) and `WithStringifyAllowDots(true)` (Go, Stringify) enables dot notation support for nested objects.

## Stringify (without and with allowDots)

JS:

```js
qs.stringify({ a: { b: "c" } }, { encode: false })
// a[b]=c

qs.stringify({ a: { b: "c" } }, { allowDots: true, encode: false })
// a.b=c
```

Go:

```go
qs.Stringify(map[string]any{"a": map[string]any{"b": "c"}}, qs.WithEncode(false))
// a[b]=c

qs.Stringify(
  map[string]any{"a": map[string]any{"b": "c"}},
  qs.WithStringifyAllowDots(true),
  qs.WithEncode(false),
)
// a.b=c
```

## Parse (without and with allowDots)

JS:

```js
qs.parse("a.b=c")
// {"a.b":"c"}

qs.parse("a.b=c", { allowDots: true })
// {"a":{"b":"c"}}
```

Go:

```go
qs.Parse("a.b=c")
// {"a.b":"c"}

qs.Parse("a.b=c", qs.WithAllowDots(true))
// {"a":{"b":"c"}}
```


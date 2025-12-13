# allowEmptyArrays / AllowEmptyArrays

The `allowEmptyArrays` option (JS) / `WithParseAllowEmptyArrays(true)` (Go, Parse) and `WithStringifyAllowEmptyArrays(true)` (Go, Stringify) controls how empty arrays are represented.

## Parse

JS:

```js
qs.parse("a[]=")
// {"a":[""]}

qs.parse("a[]=", { allowEmptyArrays: true })
// {"a":[]}
```

Go:

```go
qs.Parse("a[]=")
// {"a":[""]}

qs.Parse("a[]=", qs.WithParseAllowEmptyArrays(true))
// {"a":[]}
```

## Stringify

JS:

```js
qs.stringify({ a: [] }, { encode: false })
// (empty)

qs.stringify({ a: [] }, { allowEmptyArrays: true, encode: false })
// a[]
```

Go:

```go
qs.Stringify(map[string]any{"a": []any{}}, qs.WithStringifyEncode(false))
// (empty)

qs.Stringify(
  map[string]any{"a": []any{}},
  qs.WithStringifyAllowEmptyArrays(true),
  qs.WithStringifyEncode(false),
)
// a[]
```


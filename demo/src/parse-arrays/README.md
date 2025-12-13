# parseArrays / ParseArrays

The `parseArrays` option (JS) / `WithParseArrays(false)` (Go, Parse) disables special handling of bracket array/object syntax.

## Parse

JS:

```js
qs.parse("a[0]=b")
// {"a":["b"]}

qs.parse("a[0]=b", { parseArrays: false })
// {"a":{"0":"b"}}
```

Go:

```go
qs.Parse("a[0]=b")
// {"a":["b"]}

qs.Parse("a[0]=b", qs.WithParseArrays(false))
// {"a":{"0":"b"}}
```

## Stringify

Stringify does not have a parseArrays option; use `ArrayFormat` to choose an output shape.


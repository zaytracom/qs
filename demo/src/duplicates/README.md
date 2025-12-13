# duplicates / Duplicates

The `duplicates` option (JS) / `WithParseDuplicates(...)` (Go, Parse) controls how repeated keys are handled.

## Parse

JS:

```js
qs.parse("a=1&a=2", { duplicates: "combine" })
// {"a":["1","2"]}

qs.parse("a=1&a=2", { duplicates: "first" })
// {"a":"1"}

qs.parse("a=1&a=2", { duplicates: "last" })
// {"a":"2"}
```

Go:

```go
qs.Parse("a=1&a=2", qs.WithParseDuplicates(qs.DuplicateCombine))
// {"a":["1","2"]}

qs.Parse("a=1&a=2", qs.WithParseDuplicates(qs.DuplicateFirst))
// {"a":"1"}

qs.Parse("a=1&a=2", qs.WithParseDuplicates(qs.DuplicateLast))
// {"a":"2"}
```

## Stringify

Stringify does not have a duplicates option; it serializes the value you provide.


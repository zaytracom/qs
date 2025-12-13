# delimiter / Delimiter

The `delimiter` option (JS) / `WithParseDelimiter(...)` (Go, Parse) and `WithStringifyDelimiter(...)` (Go, Stringify) sets the separator between `key=value` pairs.

The default delimiter is `&`, but some systems use `;` or other characters. This option lets you customize the separator for both parsing and stringifying to match your requirements.

Below we also set key sorting to ensure stable output.

## Stringify (custom delimiter)

JS:

```js
qs.stringify(
  { a: "1", b: "2" },
  { delimiter: ";", encode: false, sort: (a, b) => a.localeCompare(b) }
)
// a=1;b=2
```

Go:

```go
qs.Stringify(
  map[string]any{"a": "1", "b": "2"},
  qs.WithStringifyDelimiter(";"),
  qs.WithStringifyEncode(false),
  qs.WithStringifySort(func(a, b string) bool { return a < b }),
)
// a=1;b=2
```

## Parse (custom delimiter)

JS:

```js
qs.parse("a=1;b=2", { delimiter: ";" })
// {"a":"1","b":"2"}
```

Go:

```go
qs.Parse("a=1;b=2", qs.WithParseDelimiter(";"))
// {"a":"1","b":"2"}
```

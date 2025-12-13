# charset / Charset

The `charset` option (JS) / `WithParseCharset(...)` (Go, Parse) and `WithStringifyCharset(...)` (Go, Stringify) selects UTF-8 vs ISO-8859-1 behavior.

## Parse

JS:

```js
qs.parse("a=%E9", { charset: "iso-8859-1" })
// {"a":"é"}
```

Go:

```go
qs.Parse("a=%E9", qs.WithParseCharset(qs.CharsetISO88591))
// {"a":"é"}
```

## Stringify

JS:

```js
qs.stringify({ a: "é" }, { charset: "iso-8859-1" })
// a=%E9
```

Go:

```go
qs.Stringify(map[string]any{"a": "é"}, qs.WithStringifyCharset(qs.CharsetISO88591))
// a=%E9
```


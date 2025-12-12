# arrayFormat / ArrayFormat

The `arrayFormat` option (JS) / `WithArrayFormat(...)` (Go) determines how arrays are serialized.

In all examples below, we disable URL-encoding for easier readability.

JS:

```js
qs.stringify({ a: ["b", "c"] }, { encode: false })
// a[0]=b&a[1]=c

qs.stringify({ a: ["b", "c"] }, { arrayFormat: "brackets", encode: false })
// a[]=b&a[]=c

qs.stringify({ a: ["b", "c"] }, { arrayFormat: "repeat", encode: false })
// a=b&a=c

qs.stringify({ a: ["b", "c"] }, { arrayFormat: "comma", encode: false })
// a=b,c
```

Go:

```go
qs.Stringify(map[string]any{"a": []any{"b", "c"}}, qs.WithEncode(false))
// a[0]=b&a[1]=c

qs.Stringify(map[string]any{"a": []any{"b", "c"}}, qs.WithArrayFormat(qs.ArrayFormatBrackets), qs.WithEncode(false))
// a[]=b&a[]=c

qs.Stringify(map[string]any{"a": []any{"b", "c"}}, qs.WithArrayFormat(qs.ArrayFormatRepeat), qs.WithEncode(false))
// a=b&a=c

qs.Stringify(map[string]any{"a": []any{"b", "c"}}, qs.WithArrayFormat(qs.ArrayFormatComma), qs.WithEncode(false))
// a=b,c
```


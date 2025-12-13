# sortArrayIndices / SortArrayIndices

The `sortArrayIndices` option (Go) enables sorting array indices as strings when using the `Sort` option. This is needed for compatibility with the JS `qs` library which sorts all keys including array indices when a sort function is provided.

By default, Go preserves the natural numeric order of array indices (0, 1, 2, ..., 10, 11). With `SortArrayIndices(true)`, indices are sorted lexicographically as strings (0, 1, 10, 11, 2, 3, ...), matching JS `qs` behavior with `sort`.

## Stringify (with Sort, without and with SortArrayIndices)

JS:

```js
qs.stringify({ arr: ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"] }, { encode: false, sort: (a, b) => a.localeCompare(b) })
// arr[0]=a&arr[1]=b&arr[10]=k&arr[11]=l&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j
```

Go (default - numeric order):

```go
sortAsc := func(a, b string) bool { return a < b }
qs.Stringify(
  map[string]any{"arr": []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}},
  qs.WithStringifyEncode(false),
  qs.WithStringifySort(sortAsc),
)
// arr[0]=a&arr[1]=b&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j&arr[10]=k&arr[11]=l
```

Go (with SortArrayIndices - matches JS):

```go
sortAsc := func(a, b string) bool { return a < b }
qs.Stringify(
  map[string]any{"arr": []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}},
  qs.WithStringifyEncode(false),
  qs.WithStringifySort(sortAsc),
  qs.WithStringifySortArrayIndices(true),
)
// arr[0]=a&arr[1]=b&arr[10]=k&arr[11]=l&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j
```

## When to Use

Use `WithStringifySortArrayIndices(true)` when you need:
- Byte-for-byte identical output with JS `qs` library using `sort` option
- Cross-platform testing where Go and JS outputs must match exactly
- Deterministic string output that matches JS behavior

Without this option, Go produces more semantically correct output (arrays maintain their natural order), but the string representation differs from JS `qs`.

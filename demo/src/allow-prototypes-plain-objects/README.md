# allowPrototypes / plainObjects (JS-only)

JS `qs` includes options like `allowPrototypes` and `plainObjects` as prototype-pollution mitigations for JavaScript object creation.

Go does not have prototype inheritance, so `qs/v2` does not expose these options. Keys such as `__proto__`, `constructor`, and `prototype` are treated as normal map keys.

## Parse

Go:

```go
qs.Parse("__proto__[polluted]=1")
// {"__proto__":{"polluted":"1"}}
```

## Stringify

Go:

```go
qs.Stringify(map[string]any{"__proto__": map[string]any{"polluted": "1"}}, qs.WithStringifyEncode(false))
// __proto__[polluted]=1
```


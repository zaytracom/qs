# UnmarshalBytes

`UnmarshalBytes` is a Go-only API that parses `[]byte` query strings directly into a destination value (struct/map/*any) without converting the input to a `string`.

## Parse into struct (zero-copy input)

Go:

```go
type User struct {
  Name string `query:"name"`
  Age  int    `query:"age"`
}

var u User
_ = qs.UnmarshalBytes([]byte("name=John&age=30"), &u)
// User{Name:"John", Age:30}
```

## Stringify

There is no `MarshalBytes`; use `qs.Marshal` / `qs.Stringify`.


# merror

[![GoDoc](https://godoc.org/github.com/wiggin77/merror?status.svg)](https://godoc.org/github.com/wiggin77/merror)
![Build Status](https://github.com/wiggin77/merror/actions/workflows/ci.yml/badge.svg?event=push)

Multiple Error aggregator for Go.

## Usage

```go
func foo() error {
  merr := merror.New()

  if err := DoSomething(); err != nil {
    merr.Append(err)
  }

  return merr.ErrorOrNil()
}
```

A bounded `merror` can be used to guard against memory ballooning.

```go
func bar() error {
  merr := merror.NewWithCap(10)

  for i := 0; i < 15; i++ {
    if err := DoSomething(); err != nil {
      merr.Append(err)
    }
  }

  fmt.Printf("Len: %d,  Overflow: %d", merr.Len(), merr.Overflow()) 
  // Len: 10,  Overflow: 5

  return merr.ErrorOrNil()
}
```

## errors.Is

If any of the errors appended to a `merror` match the target error passed to `errors.Is(err, target error)` then true is returned.

## errors.As

If any of the errors appended to a `merror` match the target type passed to `errors.As(err error, target any)` then true is returned and the target is set to the matching error.

## depend
A simple dependency injection tool for go.

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg)](http://godoc.org/github.com/blbergwall/depend)
[![Go Report Card](https://goreportcard.com/badge/github.com/blbergwall/depend)](https://goreportcard.com/report/github.com/blbergwall/depend)
[![License](http://img.shields.io/badge/license-mit-blue.svg)](https://github.com/blbergwall/depend/blob/master/LICENSE.txt)

### Install
```
go get github.com/blbergwall/depend
```

### Overview
Simple because it only deals with interfaces.

All dependencies are first added, then they are all built and interdependences
resolved and verified.  The result can be used to provide dependencies to
functions.

### Usage

Import the depend package
```
import "github.com/blbergwall/depend"
```

Create a new dependency builder:
```
builder := depend.New()
}
```

Assume the following interfaces and implementations that will be depended on:
```
type testInterface1 interface {
	Method1() string
}

type testInterface2 interface {
	Method2() string
}

type testStruct1 struct {
	field string
}

type testStruct2 struct {
	field string
}

func new1() testInterface1 {
	return &testStruct1{field: "testInterface1"}
}

func (ts1 *testStruct1) Method1() string {
	return ts1.field
}

func new2Consume1(i1 testInterface1) testInterface1 {
	return &testStruct1{field: "testInterface2 from 1"}
}

func (ts2 *testStruct2) Method2() string {
	return ts2.field
}
```

Then these can be added with:
```
builder.Add(new1)
builder.Add(new2Consume1)
```

Once all dependencies have been added:
```
provider, err := builder.Build()
```
This causes all added functions to be called at most once.  If any functions
return errors or any functions have dependencies that have not been added or
there are any circular references an error will be returned.

Now given a function:
```
func toProvideFor(i2 testInterface2) (string, error) {
	return i2.Method1(), nil
}
```

provider can be used to provide for it:
```
result, err := provider.ProvideFor(toProvideFor)
```
In general result will be the first result from the function and err will be
the last result from the function if it is an error.

If more than one of the same interface is added then only a slice of that
interface will be provided.

For more examples see the [tests](https://github.com/blbergwall/depend/blob/master/depend)test.go)

Also check out the [documentation](http://godoc.org/github.com/blbergwall/depend)

### Why?
Needed a dependency injection system for a project I was working on and
existing systems I found did not work quite the way I wanted.

One of the main points of dependency injection is to make code more testable.
Only injecting interfaces makes code more testable and anything can be wrapped
in an interface.  So I am hoping the interface only constraint will be a net
benefit.

Also I am fairly new to go and publishing open source code and want to learn
more about both.

### License
[MIT](https://github.com/blbergwall/depend/blob/master/LICENSE.txt)

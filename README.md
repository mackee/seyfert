# seyfert
> Template engine for Golang code.

**THIS IS A ALPHA QUALITY RELEASE. API MAY CHANGE WITHOUT NOTICE.**

## Table of Contents

* [Install](https://github.com/mackee/seyfert/blob/master/README.md#install)
* [Usage](https://github.com/mackee/seyfert/blob/master/README.md#usage)
* [Contribute](https://github.com/mackee/seyfert/blob/master/README.md#contribute)
* [License](https://github.com/mackee/seyfert/blob/master/README.md#license)

## Install

```
$ go get github.com/mackee/seyfert
```

## Usage

**Source**:
```go
package main

import (
	"github.com/mackee/seyfert"
)

var tmpl = byte(`
package main

//+seyfert
type T_ struct {
	//+expand F
}

//+seyfert
func (t T_) String() string {
	return t.Name
}

`)

func main() {
	binds := seyfert.Binds{
		"T": "Person",
	}
	fiedsSet := seyfert.FieldsSet{
		"F": seyfert.Fields{
			seyfert.Field{
				Name: "Name",
				Type: "string",
				Tag:  `json:"name"`,
			},
			seyfert.Field{
				Name: "Age",
				Type: "int",
				Tag:  `json:"age"`,
			},
		},
	}
	err := seyfert.Render(tmpl, "person.gen.go", binds, fieldsSet, "main")
	if err != nil {
		panic(err)
	}
}
```

**Generated File**:

```go
package main

//+seyfert
type Person struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

//+seyfert
func (t Person) String() string {
	return t.Name
}
```

See also: https://github.com/mackee/seyfert/blob/master/_example/genreqparser

## Contribute

PRs accepted.

## License

MIT © mackee


# go-qs

[![Go Reference](https://pkg.go.dev/badge/github.com/globocom/go-qs.svg)](https://pkg.go.dev/github.com/globocom/go-qs) [![Unit tests](https://github.com/globocom/go-qs/actions/workflows/tests.yaml/badge.svg?branch=main)](https://github.com/globocom/go-qs/actions/workflows/tests.yaml)

> Parse and decode URL query strings into nested Go data structures. Compatible with [javascript qs](npmjs.com/package/qs).

---

## Overview

`go-qs` is a **work in progress** Go library for parsing URL query strings into flexible, deeply nested Go maps and slices. It supports arrays, objects, numeric indices, and customizable parsing options, making it ideal for web applications and APIs.

## Installation

```sh
go get github.com/globocom/go-qs
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/globocom/go-qs"
)

func main() {
	result, err := goqs.Parse("user[name]=Alice&user[age]=30", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
	// Output: map[user:map[age:30 name:Alice]]
}
```

## Features

- Parse query strings into `map[string]interface{}`
- Support for arrays: `arr[]=1&arr[]=2`
- Support for nested objects: `user[name]=Alice&user[age]=30`
- Numeric indices: `a[1]=1&a[2]=2`
- Customizable parsing options (array limits, depth, charset, etc.)
- Handles empty values, custom delimiters, and more

## Options

Customize parsing behavior using `ParseOptions`:

```go
opts := &goqs.ParseOptions{
	AllowDots: true,
	ArrayLimit: 20,
	Depth: 5,
	// ...other options
}
result, err := goqs.Parse("a.b=1&a.c=2", opts)
```

See [ParseOptions](https://pkg.go.dev/github.com/globocom/go-qs#ParseOptions) for all available fields.

## Testing

Unit tests are provided in `parse_test.go` covering various scenarios:

- Empty queries
- Simple parameters
- Arrays and nested objects
- Numeric indices

Run tests with:

```sh
go test
```


## Contributing

We welcome contributions! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on pull requests and issues, including best practices for reporting bugs and comparing with the JavaScript `qs` library.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Inspiration

This project was inspired by [ljharb/qs](https://github.com/ljharb/qs), a popular query string parser for JavaScript.

---

For more details, see the [GoDoc](https://pkg.go.dev/github.com/globocom/go-qs).

Made with love ❤️ and ☕ by Backstage

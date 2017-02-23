# goforce [![Build Status](https://travis-ci.org/jpmonette/goforce.svg?branch=master)](https://travis-ci.org/jpmonette/goforce)
A Go library for Salesforce.com REST API

## Installation

```bash
$ go get -u github.com/jpmonette/goforce
```

## Quick Start

```go
package main

import (
  "github.com/jpmonette/goforce"
)

func main() {
  conn, err := Login(
    "3MVG9y6x0357Hledq10492OWNe9Ed0gthITMddp6Qjj0XIxPyrrHs4HePtL7XOuMhgHi6G0aBBl91kTXxa_Uo",
    "82559386934830682665",
    "https://login.salesforce.com",
    "standarduser@testorg.com",
    "password12345",
    "34.0",
  )

  limits, err := conn.Limits()

  fmt.Println(limits.HourlyDashboardStatuses.Remaining))
}

```

## More Information

* [Force.com REST API](https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/)
* [Force.com Tooling API](https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/intro_api_tooling.htm)
* [@jpmonette](https://twitter.com/jpmonette) on Twitter
* [LinkedIn](https://uk.linkedin.com/in/jpmonette)
* Read my personal blog [Blogue de Jean-Philippe Monette](http://blogue.jpmonette.net/) to learn more about what I do!

## License

The MIT License (MIT)

Copyright (c) 2015 Jean-Philippe Monette

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
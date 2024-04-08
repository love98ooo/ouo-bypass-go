# ouo-bypass-go

A package to bypass ouo.io and ouo.press link shortener.

## Features

- Bypass reCAPTCHA v3.
- Bypass ouo.io and ouo.press link shortener.

## Installation

``` bash
go get github.com/love98ooo/ouo-bypass-go
```

## Usage

``` go
import (
    "fmt"
    ouoBypass "github.com/love98ooo/ouo-bypass-go"
)

func main() {
    url := "https://ouo.io/xxxxxx"
    bypassedURL, err := ouoBypass.Resolve(url)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(bypassedURL)
    }
}
```

## Contributing

Issue and PR welcome.

## Acknowledgement

ouo-bypass-go is inspired by the following projects and so on:

- [ouo-bypass](https://github.com/xcscxr/ouo-bypass)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
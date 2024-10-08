fork from https://godoc.org/github.com/ankur-anand/gojtp

⚡️ A high-performance, zero allocation, dynamic JSON Threat Protection in pure Go. 🔥

JTP provides a fast way to **validate the dynamic JSON** and protect against vulnerable JSON content-level attacks (JSON
Threat Protection) based on configured properties.

**It also validates the JSON and if JSON is Invalid it will return an error.**

### What is JSON Threat Protection

JSON requests are susceptible to attacks characterized by unusual inflation of elements and nesting levels. Attackers
use recursive techniques to consume memory resources by using huge json files to overwhelm the parser and eventually
crash the service.

JSON threat protection is terms that describe the way to minimize the risk from such attacks by defining few limits on
the json structure like length and depth validation on a json, and helps protect your applications from such intrusions.

There are situations where you do not want to parse the JSON, but do want to ensure that the JSON is not going to cause
a problem. Such as an API Gateway. It would be a PAIN for the gateway to have to know all JSON schema of all services it
is protecting. There are XML validators that perform similar functions.

### Getting Started

Installing To start using gojtp, install Go and run go get:

`$ go get -u github.com/ankur-anand/gojtp`

## Performance

On linux-amd64

```
BenchmarkTestifyNoThreatInBytes-4         500000              2628 ns/op               0 B/op          0 allocs/op
```

JSON Used

```json
{
  "simple_string": "hello word",
  "targets": [
    {
      "req_per_second": 5,
      "duration_of_time": 1,
      "utf8Key": "Hello, 世界",
      "request": {
        "endpoint": "https://httpbin.org/get",
        "http_method": "GET",
        "payload": {
          "username": "ankur",
          "password": "ananad"
        },
        "array_value": [
          "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstv"
        ],
        "additional_header": [
          {
            "header_key": "uuid",
            "header_value": [
              "1",
              "2"
            ]
          }
        ]
      }
    },
    {
      "req_per_second": 10,
      "duration_of_time": 1,
      "request": {
        "endpoint": "https://httpbin.org/post",
        "http_method": "POST",
        "payload": {
          "username": "ankur",
          "password": "ananad"
        },
        "additional_header": [
          {
            "header_key": "uuid",
            "header_value": [
              "1",
              "2",
              "3",
              "4",
              "5",
              "Hello, 世界"
            ]
          }
        ]
      }
    }
  ]
}
```

### Create a verify

All the verifier Parameters are Optional

> Check Godoc for all option

Example Verify

```go
// with multiple config
_ = jj.NewJtp(WithMaxArrayElementCount(6),
WithMaxContainerDepth(7),
WithMaxObjectKeyLength(20), WithMaxStringLength(50),
)

// with single config
_ = jj.NewJtp(WithMaxStringLength(25))
```

### Errors

The JTP returns following error messages on Validation failure:

| Error Message                                             |
|---------------------------------------------------------------|
| jtp.maxStringLenReached.Max-[X]-Allowed.Found-[Y]: jtp.MalformedJSON   |
| jtp.maxArrayLenReached.Max-[X]-Allowed.Found-[Y]: jtp.MalformedJSON   |
| jtp.maxKeyLenReached.Max-[X]-Allowed.Found-[Y]: jtp.MalformedJSON |
| jtp.maxDepthReached.Max-[X]-Allowed.Found-[Y]: jtp.MalformedJSON   |
| jtp.maxEntryCountReached.Max-[X]-Allowed.Found-[Y]: jtp.MalformedJSON |
| jtp.MalformedJSON | 

## Usage Example

```go
package main

import (
	"fmt"
	"github.com/bingoohuang/ngg/jj"
)

func main() {
	json := _getTestJsonBytes()
	verifier1 := jj.NewJtp(
		WithMaxArrayLen(6),
		WithMaxDepth(7),
		WithMaxKeyLen(20), WithMaxStringLen(50),
	)
	err := verifier1.VerifyBytes(json)

	verifier2 := jj.NewJtp(WithMaxStringLen(25))
	err = verifier2.VerifyBytes(json)
	fmt.Println(err)
}
```

## Contact

Ankur Anand [@in_aanand](https://twitter.com/in_aanand)

## License

GOJTP source code is available under the MIT [License](/LICENSE).

Based on Parser from [tidwall](https://twitter.com/tidwall).
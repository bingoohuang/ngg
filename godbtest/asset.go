// Package godbtest demonstrates sql using for sql testing in golang.
package godbtest

import _ "embed"

// DemoConf is demo configuration .
//
//go:embed db.yml
var DemoConf []byte

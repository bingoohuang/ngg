package gnet

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/bingoohuang/ngg/ss"
)

var reScheme = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+-.]*://`)

type FixURI struct {
	DefaultScheme string
	DefaultHost   string
	DefaultPort   int
	Auth          string
}

func (f FixURI) Fix(uri string) (*url.URL, error) {
	f.DefaultScheme = ss.Or(f.DefaultScheme, "http")
	f.DefaultHost = ss.Or(f.DefaultHost, "127.0.0.1")

	if uri == ":" {
		uri = ":" + strconv.Itoa(f.DefaultPort)
	}

	// ex) :8080/hello or /hello or :
	if strings.HasPrefix(uri, ":") || strings.HasPrefix(uri, "/") {
		uri = f.DefaultHost + uri
	}

	// ex) example.com/hello
	if !reScheme.MatchString(uri) {
		uri = f.DefaultScheme + "://" + uri
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("parse %s failed: %s", uri, err)
	}

	u.Host = strings.TrimSuffix(u.Host, ":")
	if u.Path == "" {
		u.Path = "/"
	}

	if f.Auth != "" {
		if userpass := strings.Split(f.Auth, ":"); len(userpass) == 2 {
			u.User = url.UserPassword(userpass[0], userpass[1])
		} else {
			u.User = url.User(f.Auth)
		}
	}

	return u, nil
}

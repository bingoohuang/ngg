// Package goup contains golang upload/download server and client.
package goup

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"net/http"

	"github.com/bingoohuang/ngg/ss"
)

const bearerPrefix = "Bearer "

// Bearer returns a Handler that authenticates via Bearer Auth. Writes a http.StatusUnauthorized
// if authentication fails.
func Bearer(token string, handle http.HandlerFunc) http.HandlerFunc {
	if token == "" {
		return handle
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if a := r.Header.Get(Authorization); SecureCompare(a, bearerPrefix+token) {
			handle(w, r)
		} else {
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
		}
	}
}

// SecureCompare performs a constant time compare of two strings to limit timing attacks.
func SecureCompare(given string, actual string) bool {
	givenSha := sha512.Sum512([]byte(given))
	actualSha := sha512.Sum512([]byte(actual))

	return subtle.ConstantTimeCompare(givenSha[:], actualSha[:]) == 1
}

// BearerTokenGenerate generates a random bearer token.
func BearerTokenGenerate() string {
	b := make([]byte, 15)
	_, _ = rand.Read(b)
	return ss.Base64().EncodeBytes(b, ss.Url, ss.Raw).V1.String()
}

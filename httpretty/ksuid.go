package httpretty

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"time"
)

func Ksuid() string {
	randBuffer := [16]byte{}
	io.ReadAtLeast(rand.Reader, randBuffer[:], len(randBuffer))
	var ksuid [20]byte
	copy(ksuid[4:], randBuffer[:])

	t := time.Now()
	ts := uint32(t.Unix() - 1400000000)
	binary.BigEndian.PutUint32(ksuid[:4], ts)
	b := make([]byte, 0, 27)
	return string(fastAppendEncodeBase62(b, ksuid[:]))
}

const (
	// lexographic ordering (based on Unicode table) is 0-9A-Za-z
	base62Characters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	zeroString       = "000000000000000000000000000"
)

// This function encodes the base 62 representation of the src KSUID in binary
// form into dst.
//
// In order to support a couple of optimizations the function assumes that src
// is 20 bytes long and dst is 27 bytes long.
//
// Any unused bytes in dst will be set to the padding '0' byte.
func fastEncodeBase62(dst []byte, src []byte) {
	const srcBase = 4294967296
	const dstBase = 62

	// Split src into 5 4-byte words, this is where most of the efficiency comes
	// from because this is a O(N^2) algorithm, and we make N = N / 4 by working
	// on 32 bits at a time.
	parts := [5]uint32{
		binary.BigEndian.Uint32(src[0:4]),
		binary.BigEndian.Uint32(src[4:8]),
		binary.BigEndian.Uint32(src[8:12]),
		binary.BigEndian.Uint32(src[12:16]),
		binary.BigEndian.Uint32(src[16:20]),
	}

	n := len(dst)
	bp := parts[:]
	bq := [5]uint32{}

	for len(bp) != 0 {
		quotient := bq[:0]
		remainder := uint64(0)

		for _, c := range bp {
			value := uint64(c) + uint64(remainder)*srcBase
			digit := value / dstBase
			remainder = value % dstBase

			if len(quotient) != 0 || digit != 0 {
				quotient = append(quotient, uint32(digit))
			}
		}

		// Writes at the end of the destination buffer because we computed the
		// lowest bits first.
		n--
		dst[n] = base62Characters[remainder]
		bp = quotient
	}

	// Add padding at the head of the destination buffer for all bytes that were
	// not set.
	copy(dst[:n], zeroString)
}

// This function appends the base 62 representation of the KSUID in src to dst,
// and returns the extended byte slice.
// The result is left-padded with '0' bytes to always append 27 bytes to the
// destination buffer.
func fastAppendEncodeBase62(dst []byte, src []byte) []byte {
	dst = reserve(dst, 27)
	n := len(dst)
	fastEncodeBase62(dst[n:n+27], src)
	return dst[:n+27]
}

// Ensures that at least nbytes are available in the remaining capacity of the
// destination slice, if not, a new copy is made and returned by the function.
func reserve(dst []byte, nbytes int) []byte {
	c := cap(dst)
	n := len(dst)

	if avail := c - n; avail < nbytes {
		c *= 2
		if (c - n) < nbytes {
			c = n + nbytes
		}
		b := make([]byte, n, c)
		copy(b, dst)
		dst = b
	}

	return dst
}

package ss

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"math"
	"math/big"
	"time"
)

// from https://github.com/thanhpk/randstr

// list of default letters that can be used to make a random string when calling String
// function with no letters provided
var defLetters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandString generates a random string using only letters provided in the letters parameter
// if user ommit letters parameters, this function will use defLetters instead
func RandString(n int, letters ...string) string {
	var letterRunes []rune
	if len(letters) == 0 {
		letterRunes = defLetters
	} else {
		letterRunes = []rune(letters[0])
	}

	var bb bytes.Buffer
	bb.Grow(n)
	l := uint32(len(letterRunes))
	// on each loop, generate one random rune and append to output
	for i := 0; i < n; i++ {
		bb.WriteRune(letterRunes[binary.BigEndian.Uint32(RandBytes(4))%l])
	}
	return bb.String()
}

// RandBytes generates n random bytes.
func RandBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func RandBool() bool { return RandInt64Between(0, 1) == 0 }

func RandInt64() int64 { return RandInt64N(math.MaxInt64) }

func RandInt64N(n int64) int64 {
	v, _ := rand.Int(rander, big.NewInt(n))
	return v.Int64()
}

func RandInt64Between(min, max int64) (v int64) { return RandInt64N(max-min+1) + min }

func RandIntN(n int) int {
	v, _ := rand.Int(rander, big.NewInt(int64(n)))
	return int(v.Int64())
}

func RandInt() int { return int(RandInt32()) }

func RandIntBetween(min, max int) int { return RandIntN(max-min+1) + min }

func RandInt32N(n int) int32 {
	v, _ := rand.Int(rander, big.NewInt(int64(n)))
	return int32(v.Int64())
}

func RandInt32() int32 { return RandInt32N(math.MaxInt32) }

func RandInt32Between(min, max int) int32 { return RandInt32N(max-min+1) + int32(min) }

func RandUint64N(n int64) uint64 {
	v, _ := rand.Int(rander, big.NewInt(n))
	return v.Uint64()
}

func RandUint64() (v uint64) {
	binary.Read(rander, binary.BigEndian, &v)
	return v
}

var rander = rand.Reader // random function

func RandTime() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC)
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC)
	return RandTimeBetween(min, max)
}

func RandTimeBetween(min, max time.Time) time.Time {
	minUnit, maxUnix := min.Unix(), max.Unix()
	n, _ := rand.Int(rander, big.NewInt(maxUnix-minUnit))
	return time.Unix(n.Int64()+minUnit, 0)
}

func CopyShuffle[T any](a []T) []T {
	b := make([]T, 0, len(a))
	b = append(b, a...)
	swap := func(i, j int) { b[i], b[j] = b[j], b[i] }
	for i := len(b) - 1; i > 0; i-- {
		swap(i, RandIntN(i+1))
	}
	return b
}

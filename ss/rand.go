package ss

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"math"
	"math/big"
	"time"
)

type Rand struct{}

// from https://github.com/thanhpk/randstr

// list of default letters that can be used to make a random string when calling String
// function with no letters provided
var defLetters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// String generates a random string using only letters provided in the letters parameter
// if user ommit letters parameters, this function will use defLetters instead
func (r Rand) String(n int, letters ...string) string {
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
		bb.WriteRune(letterRunes[binary.BigEndian.Uint32(r.Bytes(4))%l])
	}
	return bb.String()
}

// Bytes generates n random bytes.
func (Rand) Bytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func (r Rand) Bool() bool { return r.Int64Between(0, 1) == 0 }

func (r Rand) Int64() int64 { return r.Int64n(math.MaxInt64) }

func (Rand) Int64n(n int64) int64 {
	v, _ := rand.Int(rander, big.NewInt(n))
	return v.Int64()
}

func (r Rand) Int64Between(min, max int64) (v int64) { return r.Int64n(max-min+1) + min }

func (Rand) Intn(n int) int {
	v, _ := rand.Int(rander, big.NewInt(int64(n)))
	return int(v.Int64())
}

func (r Rand) Int() int { return int(r.Int32()) }

func (r Rand) IntBetween(min, max int) int { return r.Intn(max-min+1) + min }

func (Rand) Int32n(n int) int32 {
	v, _ := rand.Int(rander, big.NewInt(int64(n)))
	return int32(v.Int64())
}

func (r Rand) Int32() int32 { return r.Int32n(math.MaxInt32) }

func (r Rand) Int32Between(min, max int) int32 { return r.Int32n(max-min+1) + int32(min) }

func (Rand) Uint64n(n int64) uint64 {
	v, _ := rand.Int(rander, big.NewInt(n))
	return v.Uint64()
}

func (Rand) Uint64() (v uint64) {
	binary.Read(rander, binary.BigEndian, &v)
	return v
}

var rander = rand.Reader // random function

func (r Rand) Time() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC)
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC)
	return r.TimeBetween(min, max)
}

func (Rand) TimeBetween(min, max time.Time) time.Time {
	minUnit, maxUnix := min.Unix(), max.Unix()
	n, _ := rand.Int(rander, big.NewInt(maxUnix-minUnit))
	return time.Unix(n.Int64()+minUnit, 0)
}

func CopyShuffle[T any](a []T) []T {
	b := make([]T, 0, len(a))
	b = append(b, a...)
	swap := func(i, j int) { b[i], b[j] = b[j], b[i] }
	for i := len(b) - 1; i > 0; i-- {
		swap(i, Rand{}.Intn(i+1))
	}
	return b
}

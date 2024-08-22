/*
Copyright (c) 2023 Vishal Bihani

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tsid

import (
	"sync/atomic"
	"time"
)

var Epoch int64 = 1672531200000 // 2023-01-01T00:00:00.000Z

const (
	Bytes    int32 = 8
	Chars    int32 = 13 // ToString returns a string of length 13
	Bits     int32 = 22
	Mask     int32 = 0x003fffff
	Bits1024 int32 = 10
)

var AlphabetUpper = []rune("0123456789ABCDEFGHJKMNPQRSTVWXYZ")
var AlphabetLower = []rune("0123456789abcdefghjkmnpqrstvwxyz")
var AlphabetValues []int64

func init() {
	AlphabetValues = make([]int64, 128)
	for i := 0; i < len(AlphabetValues); i++ {
		AlphabetValues[i] = -1
	}

	// Numbers
	AlphabetValues['0'] = 0x00
	AlphabetValues['1'] = 0x01
	AlphabetValues['2'] = 0x02
	AlphabetValues['3'] = 0x03
	AlphabetValues['4'] = 0x04
	AlphabetValues['5'] = 0x05
	AlphabetValues['6'] = 0x06
	AlphabetValues['7'] = 0x07
	AlphabetValues['8'] = 0x08
	AlphabetValues['9'] = 0x09

	AlphabetValues['a'] = 0x0a
	AlphabetValues['b'] = 0x0b
	AlphabetValues['c'] = 0x0c
	AlphabetValues['d'] = 0x0d
	AlphabetValues['e'] = 0x0e
	AlphabetValues['f'] = 0x0f
	AlphabetValues['g'] = 0x10
	AlphabetValues['h'] = 0x11
	AlphabetValues['j'] = 0x12
	AlphabetValues['k'] = 0x13
	AlphabetValues['m'] = 0x14
	AlphabetValues['n'] = 0x15
	AlphabetValues['p'] = 0x16
	AlphabetValues['q'] = 0x17
	AlphabetValues['r'] = 0x18
	AlphabetValues['s'] = 0x19
	AlphabetValues['t'] = 0x1a
	AlphabetValues['v'] = 0x1b
	AlphabetValues['w'] = 0x1c
	AlphabetValues['x'] = 0x1d
	AlphabetValues['y'] = 0x1e
	AlphabetValues['z'] = 0x1f

	AlphabetValues['i'] = 0x01
	AlphabetValues['l'] = 0x01
	AlphabetValues['o'] = 0x00

	AlphabetValues['A'] = 0x0a
	AlphabetValues['B'] = 0x0b
	AlphabetValues['C'] = 0x0c
	AlphabetValues['D'] = 0x0d
	AlphabetValues['E'] = 0x0e
	AlphabetValues['F'] = 0x0f
	AlphabetValues['G'] = 0x10
	AlphabetValues['H'] = 0x11
	AlphabetValues['J'] = 0x12
	AlphabetValues['K'] = 0x13
	AlphabetValues['M'] = 0x14
	AlphabetValues['N'] = 0x15
	AlphabetValues['P'] = 0x16
	AlphabetValues['Q'] = 0x17
	AlphabetValues['R'] = 0x18
	AlphabetValues['S'] = 0x19
	AlphabetValues['T'] = 0x1a
	AlphabetValues['V'] = 0x1b
	AlphabetValues['W'] = 0x1c
	AlphabetValues['X'] = 0x1d
	AlphabetValues['Y'] = 0x1e
	AlphabetValues['Z'] = 0x1f

	AlphabetValues['I'] = 0x01
	AlphabetValues['L'] = 0x01
	AlphabetValues['O'] = 0x00
}

type Tsid struct {
	Number int64
}

// New returns pointer to new tsid
func New(number int64) *Tsid {
	return &Tsid{
		Number: number,
	}
}

// Fast returns a pointer to new random tsid
func Fast() *Tsid {
	// Incrementing before using it
	cnt := atomicCounter.Add(1)

	tim := (time.Now().UnixMilli() - Epoch) << Bits
	tail := cnt & uint32(Mask)

	return New(tim | int64(tail))
}

var atomicCounter atomic.Uint32

// FromNumber returns pointer to tsid using the given number
func FromNumber(number int64) *Tsid {
	return New(number)
}

// FromBytes returns a pointer to tsid by converting the given bytes to
// number
func FromBytes(bytes []byte) *Tsid {
	var n int64 = 0
	n |= int64(bytes[0]&0xff) << 56
	n |= int64(bytes[1]&0xff) << 48
	n |= int64(bytes[2]&0xff) << 40
	n |= int64(bytes[3]&0xff) << 32
	n |= int64(bytes[4]&0xff) << 24
	n |= int64(bytes[5]&0xff) << 16
	n |= int64(bytes[6]&0xff) << 8
	n |= int64(bytes[7]) & 0xff

	return New(n)
}

// FromString returns pointer to tsid by converting the given string to
// number. It validates the string before conversion.
func FromString(str string) *Tsid {
	arr := ToRuneArray(str)

	var n int64 = 0
	n |= AlphabetValues[arr[0]] << 60
	n |= AlphabetValues[arr[1]] << 55
	n |= AlphabetValues[arr[2]] << 50
	n |= AlphabetValues[arr[3]] << 45
	n |= AlphabetValues[arr[4]] << 40
	n |= AlphabetValues[arr[5]] << 35
	n |= AlphabetValues[arr[6]] << 30
	n |= AlphabetValues[arr[7]] << 25
	n |= AlphabetValues[arr[8]] << 20
	n |= AlphabetValues[arr[9]] << 15
	n |= AlphabetValues[arr[10]] << 10
	n |= AlphabetValues[arr[11]] << 5
	n |= AlphabetValues[arr[12]]

	return New(n)
}

// ToRuneArray converts the given string to rune array. It also performs
// validations on the rune array
func ToRuneArray(str string) []rune {
	arr := []rune(str)

	if !IsValidRuneArray(arr) {
		return nil // TODO: Throw error
	}
	return arr
}

// IsValidRuneArray validates the rune array.
func IsValidRuneArray(arr []rune) bool {

	if arr == nil || len(arr) != int(Chars) {
		return false
	}

	if (AlphabetValues[arr[0]] & 0b10000) != 0 {
		return false
	}

	for i := 0; i < len(arr); i++ {
		if AlphabetValues[arr[i]] == -1 {
			return false
		}
	}
	return true
}

// ToNumber returns the numerical component of the tsid
func (t *Tsid) ToNumber() int64 {
	return t.Number
}

// ToBytes converts the number to bytes and returns the byte array
func (t *Tsid) ToBytes() []byte {
	b := make([]byte, Bytes)

	b[0] = byte(uint64(t.Number) >> 56)
	b[1] = byte(uint64(t.Number) >> 48)
	b[2] = byte(uint64(t.Number) >> 40)
	b[3] = byte(uint64(t.Number) >> 32)
	b[4] = byte(uint64(t.Number) >> 24)
	b[5] = byte(uint64(t.Number) >> 16)
	b[6] = byte(uint64(t.Number) >> 8)
	b[7] = byte(t.Number)

	return b
}

// ToString converts the number to a canonical string.
// The output is 13 characters long and only contains characters from
// Crockford's base32 alphabets
func (t *Tsid) ToString() string {
	return t.ToStringWithAlphabets(AlphabetUpper)
}

// ToLower converts the number to a canonical string in lower case.
// The output is 13 characters long and only contains characters from
// Crockford's base32 alphabets
func (t *Tsid) ToLower() string {
	return t.ToStringWithAlphabets(AlphabetLower)
}

// ToStringWithAlphabets converts the number to string using the given alphabets and returns it
func (t *Tsid) ToStringWithAlphabets(alphabets []rune) string {
	c := make([]rune, Chars)

	c[0] = alphabets[((uint64(t.Number) >> 60) & 0b11111)]
	c[1] = alphabets[((uint64(t.Number) >> 55) & 0b11111)]
	c[2] = alphabets[((uint64(t.Number) >> 50) & 0b11111)]
	c[3] = alphabets[((uint64(t.Number) >> 45) & 0b11111)]
	c[4] = alphabets[((uint64(t.Number) >> 40) & 0b11111)]
	c[5] = alphabets[((uint64(t.Number) >> 35) & 0b11111)]
	c[6] = alphabets[((uint64(t.Number) >> 30) & 0b11111)]
	c[7] = alphabets[((uint64(t.Number) >> 25) & 0b11111)]
	c[8] = alphabets[((uint64(t.Number) >> 20) & 0b11111)]
	c[9] = alphabets[((uint64(t.Number) >> 15) & 0b11111)]
	c[10] = alphabets[((uint64(t.Number) >> 10) & 0b11111)]
	c[11] = alphabets[((uint64(t.Number) >> 5) & 0b11111)]
	c[12] = alphabets[(uint64(t.Number) & 0b11111)]

	return string(c)
}

// IsValid checks if the given tsid string is valid or not
func (t *Tsid) IsValid(str string) bool {
	return len(str) != 0 && IsValidRuneArray([]rune(str))
}

// GetRandom returns random component (node + counter) of the tsid
func (t *Tsid) GetRandom() int64 {
	return t.Number & int64(Mask)
}

// GetUnixMillis returns time of creation in millis since 1970-01-01
func (t *Tsid) GetUnixMillis() int64 {
	return t.getTime() + Epoch
}

// GetUnixMillisEpoch returns time of creation in millis since 1970-01-01
func (t *Tsid) GetUnixMillisEpoch(epoch int64) int64 {
	return t.getTime() + epoch
}

// getTime returns the time component
func (t *Tsid) getTime() int64 {
	return int64(uint64(t.Number) >> Bits)
}

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
	cryptorand "crypto/rand"
	"math"
	"math/big"
	mathrand "math/rand"
	"time"
)

const (
	ByteSize       = 8
	IntegerSize32  = 32
	IntegerBytes32 = 4
)

type Random interface {
	NextInt() (int32, error)
	NextBytes(length int32) ([]byte, error)
}

type IntRandom struct {
	supplier IntSupplier
}

func NewIntRandom(intSupplier IntSupplier) *IntRandom {
	return &IntRandom{
		supplier: intSupplier,
	}
}

func (i *IntRandom) NextInt() (int32, error) {
	return i.supplier.GetInt()
}

func (i *IntRandom) NextBytes(length int32) ([]byte, error) {
	bytes := make([]byte, length)

	shift := 0
	var random int32 = 0
	var err error = nil

	for j := 0; j < int(length); j++ {
		if shift < ByteSize {
			shift = IntegerSize32

			// generate random value
			random, err = i.supplier.GetInt()
			if err != nil {
				return nil, err
			}
		}
		shift -= ByteSize
		bytes[j] = byte(uint32(random) >> shift)
	}

	return bytes, nil
}

type ByteRandom struct {
	byteSupplier ByteSupplier
}

func NewByteRandom(supplier ByteSupplier) *ByteRandom {
	return &ByteRandom{
		byteSupplier: supplier,
	}
}

func (i *ByteRandom) NextInt() (int32, error) {
	var number int32 = 0

	bytes, err := i.byteSupplier.GetBytes(IntegerSize32)
	if err != nil {
		return number, err
	}

	for j := 0; j < IntegerBytes32; j++ {
		number = int32(byte(number<<ByteSize) | (bytes[j] & 0xff))
	}
	return number, nil
}

func (i *ByteRandom) NextBytes(length int32) ([]byte, error) {
	return i.byteSupplier.GetBytes(length)
}

type IntSupplier interface {
	GetInt() (int32, error)
}

type IntSupplierFunc func() (int32, error)

func (f IntSupplierFunc) GetInt() (int32, error) { return f() }

type ByteSupplier interface {
	GetBytes(length int32) ([]byte, error)
}

type ByteSupplierFunc func(length int32) ([]byte, error)

func (f ByteSupplierFunc) GetBytes(length int32) ([]byte, error) {
	return f(length)
}

type MathRandomSupplier struct {
}

func NewMathRandomSupplier() *MathRandomSupplier {
	return &MathRandomSupplier{}
}

func (i *MathRandomSupplier) GetInt() (int32, error) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	return r.Int31(), nil
}

func (i *MathRandomSupplier) GetBytes(length int32) ([]byte, error) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))

	bytes := make([]byte, length)
	_, err := r.Read(bytes)
	return bytes, err
}

type CryptoRandomSupplier struct {
}

func NewCryptoRandomSupplier() *CryptoRandomSupplier {
	return &CryptoRandomSupplier{}
}

func (i *CryptoRandomSupplier) GetInt() (int32, error) {
	r, err := cryptorand.Int(cryptorand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		return 0, err
	}
	return int32(r.Int64()), nil
}

func (i *CryptoRandomSupplier) GetBytes(length int32) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := cryptorand.Read(bytes)

	return bytes, err
}

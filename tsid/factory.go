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
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	// Lock will be used to control access for creating Factory instance
	lock sync.Mutex

	// Lock will be used to control synchronize access to random value generator
	rLock sync.Mutex

	// Only a single instance of Factory will be used per node
	factoryInstance *Factory
)

type Clock interface {
	UnixMilli() int64
}

// Factory is a singleton that
// should be used to generate random Tsid
type Factory struct {
	node        int32
	nodeBits    int32
	nodeMask    int32
	counter     int32
	counterBits int32
	counterMask int32
	lastTime    int64
	customEpoch int64
	clock       Clock
	random      Random
	randomBytes int32
}

func newFactory(builder *Builder) (*Factory, error) {
	// properties from builder
	fact := &Factory{
		customEpoch: builder.GetEpoch(),
		clock:       builder.GetClock(),
		random:      builder.GetRandom(),
	}

	// get node bits
	nodeBits, err := builder.GetNodeBits()
	if err != nil {
		log.Print(err.Error())
		return nil, errors.New("failed to initialize tsid factory")
	}
	fact.nodeBits = nodeBits

	// properties to be calculated
	fact.counterBits = Bits - builder.nodeBits
	fact.counterMask = int32(uint32(Mask) >> builder.nodeBits)
	fact.nodeMask = int32(uint32(Mask) >> fact.counterBits)

	fact.randomBytes = ((fact.counterBits - 1) / 8) + 1

	// get node id
	node, err := builder.GetNode()
	if err != nil {
		log.Print(err.Error())
		return nil, errors.New("failed to initialize tsid factory")
	}
	fact.node = node & fact.nodeMask

	fact.lastTime = fact.clock.UnixMilli()
	counter, err := fact.getRandomValue()
	if err != nil {
		return nil, err
	}

	fact.counter = counter
	return fact, nil
}

// Generate will return a tsid with random number
func (factory *Factory) Generate() (*Tsid, error) {
	tim, err := factory.getTime()
	if err != nil {
		return nil, err
	}

	tim = tim << Bits
	node := factory.node << factory.counterBits
	counter := factory.counter & factory.counterMask

	number := tim | int64(node) | int64(counter)
	return New(number), nil
}

func (factory *Factory) getTime() (int64, error) {
	milli := factory.clock.UnixMilli()
	if milli <= factory.lastTime {
		factory.counter++
		carry := uint32(factory.counter) >> factory.counterBits
		factory.counter = factory.counter & factory.counterMask
		milli = factory.lastTime + int64(carry)

	} else {
		value, err := factory.getRandomValue()
		if err != nil {
			return 0, err
		}
		factory.counter = value
	}
	factory.lastTime = milli
	return milli - factory.customEpoch, nil
}

func (factory *Factory) getRandomValue() (int32, error) {
	switch factory.random.(type) {
	case *ByteRandom:
		rLock.Lock()
		bytes, err := factory.random.NextBytes(factory.randomBytes)
		rLock.Unlock()

		if err != nil {
			return 0, err
		}

		switch len(bytes) {
		case 1:
			return int32((bytes[0] & 0xff) & byte(factory.counterMask)), nil
		case 2:
			return ((int32(bytes[0]&0xff) << 8) | int32(bytes[1]&0xff)) & factory.counterMask, nil
		case 3:
			return ((int32(bytes[0]&0xff) << 16) | (int32(bytes[1]&0xff) << 8) |
				int32(bytes[2]&0xff)) & factory.counterMask, nil
		}
	case *IntRandom:
		rLock.Lock()
		value, err := factory.random.NextInt()
		rLock.Unlock()
		if err != nil {
			return 0, err
		}

		return value & factory.counterMask, nil
	}

	return 0, errors.New("invalid random")
}

type Builder struct {
	node     int32
	nodeBits int32
	epoch    int64
	clock    Clock
	random   Random
}

// NewBuilder should be used to get instance of factory
func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) WithNode(node int32) *Builder {
	b.node = node
	return b
}

func (b *Builder) WithNodeBits(nodeBits int32) *Builder {
	b.nodeBits = nodeBits
	return b
}

func (b *Builder) WithEpoch(epoch int64) *Builder {
	b.epoch = epoch
	return b
}

func (b *Builder) WithClock(clock Clock) *Builder {
	b.clock = clock
	return b
}

func (b *Builder) WithRandom(random Random) *Builder {
	b.random = random
	return b
}

// GetNode returns the provided node id. Default is zero.
func (b *Builder) GetNode() (int32, error) {
	if b.nodeBits <= 0 {
		return 0, nil
	}
	maxNode := int32(1<<b.nodeBits) - 1

	if b.node < 0 || b.node > maxNode {
		err := fmt.Sprintf("node id out of range [0, %d]: %d", maxNode, b.node)
		return 0, errors.New(err)
	}
	return b.node, nil
}

// GetNodeBits returns the provided node bits. Default is zero.
// Range: [0, 20]
func (b *Builder) GetNodeBits() (int32, error) {
	maxNode := 20

	if b.nodeBits < 0 || b.nodeBits > 20 {
		err := fmt.Sprintf("node bits out of range [0, %d]: %d", maxNode, b.nodeBits)
		return 0, errors.New(err)
	}
	return b.nodeBits, nil
}

func (b *Builder) GetClock() Clock {
	if b.clock == nil {
		b.clock = time.Now().UTC()
	}
	return b.clock
}

func (b *Builder) GetRandom() Random {
	if b.random == nil {
		randomSupplier := NewMathRandomSupplier()
		b.random = NewIntRandom(randomSupplier)
	}

	return b.random
}

func (b *Builder) GetEpoch() int64 {
	if b.epoch == 0 {
		b.epoch = Epoch
	}
	return b.epoch
}

func (b *Builder) Build() (*Factory, error) {
	if factoryInstance != nil {
		return factoryInstance, nil
	}

	lock.Lock()
	defer lock.Unlock()

	var err error = nil
	if factoryInstance == nil {
		factoryInstance, err = newFactory(b)
		if err != nil {
			return nil, err
		}
	}
	return factoryInstance, nil
}

func (b *Builder) New() (*Factory, error) {
	return newFactory(b)
}

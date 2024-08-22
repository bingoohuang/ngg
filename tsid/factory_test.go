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

package tsid_test

import (
	"testing"
	"time"

	"github.com/bingoohuang/ngg/tsid"
	"github.com/stretchr/testify/assert"
)

func Test_WithNode(t *testing.T) {

	t.Run("given node id greater than zero tsid should contain correct node id", func(t *testing.T) {
		for i := 0; i < 20; i++ {

			nodeBits := tsid.Bits1024
			shift := tsid.Bits - nodeBits
			mask := (1 << nodeBits) - 1

			node := int32(500 & mask)
			instance, _ := tsid.NewBuilder().
				WithNode(node).
				WithNodeBits(nodeBits).
				New()
			assert.NotNil(t, instance)

			tsid, _ := instance.Generate()
			assert.NotNil(t, tsid)

			actualNode := int32(uint32(tsid.GetRandom())>>shift) & int32(mask)
			assert.Equal(t, node, actualNode, "Node id does not match the provided id")
		}
	})

	t.Run("should use default node id as zero", func(t *testing.T) {
		for i := 0; i < 20; i++ {

			nodeBits := tsid.Bits1024
			shift := tsid.Bits - nodeBits
			mask := (1 << nodeBits) - 1

			instance, _ := tsid.NewBuilder().
				WithNodeBits(nodeBits).
				New()
			assert.NotNil(t, instance)

			tsid, _ := instance.Generate()
			assert.NotNil(t, tsid)

			actualNode := int32(uint32(tsid.GetRandom())>>shift) & int32(mask)
			assert.Zero(t, actualNode, "Node id does not match the default id")
		}
	})
}

func Test_WithNodeBits(t *testing.T) {

	t.Run("given node bits should use correct node bits in tsid", func(t *testing.T) {
		// possible node bits are from [0, 20]. testing all
		for i := 0; i <= 20; i++ {

			nodeBits := int32(i)
			shift := tsid.Bits - nodeBits
			mask := (1 << nodeBits) - 1

			node := int32(500 & mask)
			instance, _ := tsid.NewBuilder().
				WithNode(node).
				WithNodeBits(nodeBits).
				New()
			assert.NotNil(t, instance)

			tsid, _ := instance.Generate()
			assert.NotNil(t, tsid)

			actualNode := int32(uint32(tsid.GetRandom())>>shift) & int32(mask)
			assert.Equal(t, node, actualNode, "Node id does not match the provided id")
		}
	})

	t.Run("should use default node bits in tsid when not provided", func(t *testing.T) {
		defaultNodeBits := 0
		for i := 0; i <= 20; i++ {

			shift := tsid.Bits - int32(defaultNodeBits)
			mask := (1 << defaultNodeBits) - 1

			node := int32(500 & mask)
			instance, _ := tsid.NewBuilder().
				WithNode(node).
				New()
			assert.NotNil(t, instance)

			tsid, _ := instance.Generate()
			assert.NotNil(t, tsid)

			actualNode := int32(uint32(tsid.GetRandom())>>shift) & int32(mask)
			assert.Zero(t, actualNode, "Node id does not match the default id")
		}
	})
}

func Test_WithRandom(t *testing.T) {

	t.Run("given random should not return error", func(t *testing.T) {

		supplier := tsid.NewMathRandomSupplier()
		random := tsid.NewIntRandom(supplier)

		instance, _ := tsid.NewBuilder().
			WithRandom(random).
			New()
		assert.NotNil(t, instance)

		_, err := instance.Generate()
		assert.Nil(t, err)
	})

	t.Run("should use default random when not provided", func(t *testing.T) {

		instance, _ := tsid.NewBuilder().
			New()
		assert.NotNil(t, instance)

		_, err := instance.Generate()
		assert.Nil(t, err)
	})
}

func Test_Generate(t *testing.T) {

	t.Run("given clock when clock drifts time should not decrease", func(t *testing.T) {

		var diff int64 = 10000
		epoch := time.Now().UnixMilli()

		clock := &MockClock{
			millis: []int64{-1, epoch, epoch + 0, epoch + 1, epoch + 2, epoch + 3 - diff, epoch + 4 - diff, epoch + 5},
		}

		intRandom := tsid.NewIntRandom(tsid.IntSupplierFunc(func() (int32, error) {
			return 0, nil
		}))

		instance, _ := tsid.NewBuilder().
			WithClock(clock).
			WithRandom(intRandom).
			New()
		assert.NotNil(t, instance)

		// Generate and record the time component
		tsid1, _ := instance.Generate()
		ms1 := tsid1.GetUnixMillis() // time

		tsid2, _ := instance.Generate()
		ms2 := tsid2.GetUnixMillis() // time + 0

		tsid3, _ := instance.Generate()
		ms3 := tsid3.GetUnixMillis() // time + 1

		tsid4, _ := instance.Generate()
		ms4 := tsid4.GetUnixMillis() // time + 2

		tsid5, _ := instance.Generate()
		ms5 := tsid5.GetUnixMillis() // time + 3 - diff

		tsid6, _ := instance.Generate()
		ms6 := tsid6.GetUnixMillis() // time + 4 - diff

		tsid7, _ := instance.Generate()
		ms7 := tsid7.GetUnixMillis() // time + 5

		assert.Equal(t, ms1+0, ms2) // clock repeats
		assert.Equal(t, ms1+1, ms3) // clock advanced
		assert.Equal(t, ms1+2, ms4) // clock advanced
		assert.Equal(t, ms1+2, ms5) // clock drift, time component should not decrease
		assert.Equal(t, ms1+2, ms6) // clock drift, time component should not decrease
		assert.Equal(t, ms1+5, ms7) // clock advanced

	})

	t.Run("given clock when encounters leap second time should not decrease", func(t *testing.T) {

		second := time.Now().Second()
		leapSecond := second - 1

		clock := &MockClock{
			millis: []int64{-1, int64(second * 1000), int64(leapSecond * 1000)},
		}

		intRandom := tsid.NewIntRandom(tsid.IntSupplierFunc(func() (int32, error) {
			return 0, nil
		}))

		instance, _ := tsid.NewBuilder().
			WithClock(clock).
			WithRandom(intRandom).
			New()
		assert.NotNil(t, instance)

		// Generate and record the time component
		tsid1, _ := instance.Generate()
		ms1 := tsid1.GetUnixMillis() // second

		tsid2, _ := instance.Generate()
		ms2 := tsid2.GetUnixMillis() // leap second

		assert.Equal(t, ms1, ms2) // time component should not decrease due to leap second
	})
}

type MockClock struct {
	index  int
	millis []int64
}

func (c *MockClock) UnixMilli() int64 {
	millis := c.millis[c.index]
	c.index++

	return millis
}

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
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/tsid"
	"github.com/stretchr/testify/assert"
)

const (
	LoopMax = 1000
)

func Test_GetUnixMillis(t *testing.T) {

	t.Run("should return correct time", func(t *testing.T) {
		start := time.Now().UnixMilli()

		instance, _ := tsid.NewBuilder().New()
		assert.NotNil(t, instance)

		tsid, _ := instance.Generate()
		assert.NotNil(t, tsid)

		middle := tsid.GetUnixMillis()
		end := time.Now().UnixMilli()

		if middle < start || (middle > end) {
			t.Fail()
		}
	})

	t.Run("given custom time should return correct time", func(t *testing.T) {
		bound := math.Pow(2, 42)

		for i := 0; i < LoopMax; i++ {

			// generate random value
			random := rand.New(rand.NewSource(time.Now().UnixNano())).
				Int63n(int64(bound))

			// ensuring date is generated after TSID_EPOCH
			millis := random + tsid.Epoch
			time := time.UnixMilli(millis)

			// int random supplier func
			intRandomSupplierFunc := func() (int32, error) {
				return 0, nil
			}

			intRandom := tsid.NewIntRandom(tsid.IntSupplierFunc(intRandomSupplierFunc))

			instance, _ := tsid.NewBuilder().
				WithClock(time).
				WithRandom(intRandom).
				New()
			assert.NotNil(t, instance)

			tsid, _ := instance.Generate()
			assert.NotNil(t, tsid)

			result := tsid.GetUnixMillis()
			assert.Equal(t, millis, result)
		}
	})

	t.Run("given custom epoch should return correct time", func(t *testing.T) {

		epoch := time.Date(1984, time.January, 1, 0, 0, 0, 0, time.UTC).
			UnixMilli()

		start := time.Now().UnixMilli()

		instance, _ := tsid.NewBuilder().
			WithEpoch(epoch).
			New()
		assert.NotNil(t, instance)

		tsid, _ := instance.Generate()
		assert.NotNil(t, tsid)

		middle := tsid.GetUnixMillisEpoch(epoch)
		end := time.Now().UnixMilli()

		if middle < start || (middle > end) {
			t.Fail()
		}
	})
}

func TestCollision(t *testing.T) {

	t.Run("one goroutine per node", func(t *testing.T) {

		// One goroutine demonstrates one node
		goroutineCount := 10
		iterationCount := 100_000

		var collisionCounter atomic.Uint32
		var tsidMap sync.Map

		wg := &sync.WaitGroup{}

		for i := 0; i < goroutineCount; i++ {

			nodeId := i

			wg.Add(1)
			go func(nodeId int32, iterationCount int32, collisionCounter *atomic.Uint32,
				tsidMap *sync.Map, wg *sync.WaitGroup) {
				defer wg.Done()

				fact, err := tsid.NewBuilder().
					WithNode(nodeId).
					Build()
				assert.Nil(t, err)

				for j := 0; j < int(iterationCount); j++ {
					tsid, err := fact.Generate()
					assert.Nil(t, err)

					// check if this tsid was already generated
					if _, ok := tsidMap.Load(tsid); !ok {

						// not present, store it
						tsidMap.Store(tsid, (nodeId*iterationCount)+int32(j))
						continue
					}

					// collision detected, increment counter and break out
					collisionCounter.Add(1)
					break
				}

			}(int32(nodeId), int32(iterationCount), &collisionCounter, &tsidMap, wg)

		}

		// Wait for all goroutines to complete
		wg.Wait()

		assert.Zero(t, collisionCounter.Load(), 0, "Collision detected")
	})

	t.Run("multiple goroutines per node", func(t *testing.T) {

		node := 1
		nodeBit := 1
		goroutineCount := 10
		iterationCount := 200_000

		var collisionCounter atomic.Uint32
		var tsidMap sync.Map

		wg := &sync.WaitGroup{}

		for i := 0; i < goroutineCount; i++ {

			wg.Add(1)
			go func(nodeId int32, nodeBit int32, iterationCount int32, collisionCounter *atomic.Uint32,
				tsidMap *sync.Map, wg *sync.WaitGroup) {
				defer wg.Done()

				fact, err := tsid.NewBuilder().
					WithNodeBits(nodeBit).
					Build()
				assert.Nil(t, err)

				for j := 0; j < int(iterationCount); j++ {
					tsid, err := fact.Generate()
					assert.Nil(t, err)

					// check if this tsid was already generated
					if _, ok := tsidMap.Load(tsid); !ok {

						// not present, store it
						tsidMap.Store(tsid, (nodeId*iterationCount)+int32(j))
						continue
					}

					// collision detected, increment counter and break out
					collisionCounter.Add(1)
					break
				}

			}(int32(node), int32(nodeBit), int32(iterationCount), &collisionCounter, &tsidMap, wg)

		}

		// Wait for all goroutines to complete
		wg.Wait()

		assert.Zero(t, collisionCounter.Load(), 0, "Collision detected")
	})
}

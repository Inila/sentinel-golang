// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hotspot

import (
	"fmt"
	"math"
	"testing"

	"github.com/alibaba/sentinel-golang/core/hotspot/cache"
	"github.com/stretchr/testify/assert"
)

func Test_tcGenFuncMap(t *testing.T) {
	t.Run("Test_tcGenFuncMap_withoutMetric", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific[100] = 100
		r1 := &Rule{
			ID:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110.0,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     specific,
		}
		generator, supported := tcGenFuncMap[r1.ControlBehavior]
		assert.True(t, supported && generator != nil)
		tc := generator(r1, nil)
		assert.True(t, tc.BoundMetric() != nil && tc.BoundRule() == r1 && tc.BoundParamIndex() == 0)
		rejectTC := tc.(*rejectTrafficShapingController)
		if rejectTC == nil {
			t.Fatal("nil traffic shaping controller")
		}
		assert.True(t, rejectTC.res == r1.Resource && rejectTC.metricType == r1.MetricType && rejectTC.paramIndex == r1.ParamIndex && rejectTC.burstCount == r1.BurstCount)
		assert.True(t, rejectTC.threshold == r1.Threshold && rejectTC.durationInSec == r1.DurationInSec)
	})

	t.Run("Test_tcGenFuncMap_withMetric", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific[100] = 100
		r1 := &Rule{
			ID:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110.0,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     specific,
		}
		generator, supported := tcGenFuncMap[r1.ControlBehavior]
		assert.True(t, supported && generator != nil)

		size := int(math.Min(float64(ParamsMaxCapacity), float64(ParamsCapacityBase*r1.DurationInSec)))
		if size <= 0 {
			size = ParamsMaxCapacity
		}
		metric := &ParamsMetric{
			RuleTimeCounter:    cache.NewLRUCacheMap(size),
			RuleTokenCounter:   cache.NewLRUCacheMap(size),
			ConcurrencyCounter: cache.NewLRUCacheMap(ConcurrencyMaxCount),
		}

		tc := generator(r1, metric)
		assert.True(t, tc.BoundMetric() != nil && tc.BoundRule() == r1 && tc.BoundParamIndex() == 0)
		rejectTC := tc.(*rejectTrafficShapingController)
		if rejectTC == nil {
			t.Fatal("nil traffic shaping controller")
		}
		assert.True(t, rejectTC.metric == metric)
		assert.True(t, rejectTC.res == r1.Resource && rejectTC.metricType == r1.MetricType && rejectTC.paramIndex == r1.ParamIndex && rejectTC.burstCount == r1.BurstCount)
		assert.True(t, rejectTC.threshold == r1.Threshold && rejectTC.durationInSec == r1.DurationInSec)

	})
}

func Test_IsValidRule(t *testing.T) {
	t.Run("Test_IsValidRule", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific[100] = 100
		r1 := &Rule{
			ID:                "abc",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110.0,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     specific,
		}
		assert.True(t, IsValidRule(r1) == nil)
	})

	t.Run("Test_InValidRule", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific[100] = 100
		r1 := &Rule{
			ID:                "",
			Resource:          "",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110.0,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     specific,
		}
		assert.True(t, IsValidRule(r1) != nil)
	})

	t.Run("Test_InValidRule2", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific[100] = 100
		specific["100"] = 100
		r1 := &Rule{
			ID:                "",
			Resource:          "abc",
			MetricType:        Concurrency,
			ControlBehavior:   Reject,
			ParamIndex:        0,
			Threshold:         110,
			MaxQueueingTimeMs: 0,
			BurstCount:        10,
			DurationInSec:     1,
			SpecificItems:     specific,
		}
		assert.True(t, IsValidRule(r1) == nil)
	})
}

func Test_onRuleUpdate(t *testing.T) {
	tcMap = make(trafficControllerMap)

	specific := make(map[interface{}]int64)
	specific["sss"] = 1
	specific["123"] = 3
	r1 := &Rule{
		ID:                "1",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Reject,
		ParamIndex:        0,
		Threshold:         100.0,
		MaxQueueingTimeMs: 0,
		BurstCount:        10,
		DurationInSec:     1,
		SpecificItems:     specific,
	}

	specific2 := make(map[interface{}]int64)
	specific2["sss"] = 1
	specific2["123"] = 3
	r2 := &Rule{
		ID:                "2",
		Resource:          "abc",
		MetricType:        QPS,
		ControlBehavior:   Throttling,
		ParamIndex:        1,
		Threshold:         100.0,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     specific2,
	}

	specific3 := make(map[interface{}]int64)
	specific3["sss"] = 1
	specific3["123"] = 3
	r3 := &Rule{
		ID:                "3",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Throttling,
		ParamIndex:        2,
		Threshold:         100,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     specific3,
	}

	r4 := &Rule{
		ID:                "4",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Throttling,
		ParamIndex:        2,
		Threshold:         100.0,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     2,
		SpecificItems:     specific3,
	}

	updated, err := LoadRules([]*Rule{r1, r2, r3, r4})
	if !updated || err != nil {
		t.Errorf("Fail to prepare data, err: %+v", err)
	}
	assert.True(t, len(tcMap["abc"]) == 4)

	r21 := &Rule{
		ID:                "21",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Reject,
		ParamIndex:        0,
		Threshold:         100.0,
		MaxQueueingTimeMs: 0,
		BurstCount:        10,
		DurationInSec:     1,
		SpecificItems:     specific,
	}
	r22 := &Rule{
		ID:                "22",
		Resource:          "abc",
		MetricType:        QPS,
		ControlBehavior:   Throttling,
		ParamIndex:        1,
		Threshold:         101.0,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     1,
		SpecificItems:     specific2,
	}
	r23 := &Rule{
		ID:                "23",
		Resource:          "abc",
		MetricType:        Concurrency,
		ControlBehavior:   Throttling,
		ParamIndex:        2,
		Threshold:         100.0,
		MaxQueueingTimeMs: 20,
		BurstCount:        0,
		DurationInSec:     12,
		SpecificItems:     specific3,
	}

	oldTc1Ptr := tcMap["abc"][0]
	oldTc2Ptr := tcMap["abc"][1]
	oldTc3Ptr := tcMap["abc"][2]
	oldTc4Ptr := tcMap["abc"][3]
	oldTc1PtrAddr := fmt.Sprintf("%p", oldTc1Ptr)
	oldTc2PtrAddr := fmt.Sprintf("%p", oldTc2Ptr)
	oldTc3PtrAddr := fmt.Sprintf("%p", oldTc3Ptr)
	oldTc4PtrAddr := fmt.Sprintf("%p", oldTc4Ptr)
	fmt.Println(oldTc1PtrAddr)
	fmt.Println(oldTc2PtrAddr)
	fmt.Println(oldTc3PtrAddr)
	fmt.Println(oldTc4PtrAddr)
	oldTc2MetricPtrAddr := fmt.Sprintf("%p", tcMap["abc"][1].BoundMetric())
	fmt.Println("oldTc2MetricPtr:", oldTc2MetricPtrAddr)

	err = onRuleUpdate([]*Rule{r21, r22, r23})
	assert.True(t, err == nil)
	assert.True(t, len(tcMap) == 1)
	abcTcs := tcMap["abc"]
	assert.True(t, len(abcTcs) == 3)
	newTc1Ptr := abcTcs[0]
	newTc2Ptr := abcTcs[1]
	newTc3Ptr := abcTcs[2]
	newTc1PtrAddr := fmt.Sprintf("%p", newTc1Ptr)
	newTc2PtrAddr := fmt.Sprintf("%p", newTc2Ptr)
	newTc3PtrAddr := fmt.Sprintf("%p", newTc3Ptr)
	fmt.Println(newTc1PtrAddr)
	fmt.Println(newTc2PtrAddr)
	fmt.Println(newTc3PtrAddr)
	newTc2MetricPtrAddr := fmt.Sprintf("%p", newTc2Ptr.BoundMetric())
	fmt.Println("newTc2MetricPtrAddr:", newTc2MetricPtrAddr)
	assert.True(t, newTc1PtrAddr == oldTc1PtrAddr && newTc2MetricPtrAddr == oldTc2MetricPtrAddr)
	assert.True(t, abcTcs[0].BoundRule() == r1 && abcTcs[0] == oldTc1Ptr)
	assert.True(t, abcTcs[1].BoundMetric() == oldTc2Ptr.BoundMetric())

	tcMap = make(trafficControllerMap)
}

func TestLoadRules(t *testing.T) {
	t.Run("loadSameRules", func(t *testing.T) {
		specific := make(map[interface{}]int64)
		specific["sss"] = 1
		specific["123"] = 3

		_, err := LoadRules([]*Rule{
			{
				ID:                "1",
				Resource:          "abc",
				MetricType:        Concurrency,
				ControlBehavior:   Reject,
				ParamIndex:        0,
				Threshold:         100.0,
				MaxQueueingTimeMs: 0,
				BurstCount:        10,
				DurationInSec:     1,
				SpecificItems:     specific,
			},
		})
		assert.Nil(t, err)
		ok, err := LoadRules([]*Rule{
			{
				ID:                "1",
				Resource:          "abc",
				MetricType:        Concurrency,
				ControlBehavior:   Reject,
				ParamIndex:        0,
				Threshold:         100.0,
				MaxQueueingTimeMs: 0,
				BurstCount:        10,
				DurationInSec:     1,
				SpecificItems:     specific,
			},
		})
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

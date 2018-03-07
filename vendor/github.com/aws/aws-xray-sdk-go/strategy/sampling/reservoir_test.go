// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const Interval = 100

func takeOverTime(r *Reservoir, millis int) int {
	taken := 0
	for i := 0; i < millis/Interval; i++ {
		if r.Take() {
			taken++
		}
		time.Sleep(Interval * time.Millisecond)
	}
	return taken
}

const TestDuration = 1500

func TestOnePerSecond(t *testing.T) {
	per := 1
	res, err := NewReservoir(uint64(per))
	assert.NoError(t, err)
	taken := takeOverTime(res, TestDuration)
	assert.True(t, int(math.Ceil(TestDuration/1000.0)) <= taken)
	assert.True(t, int(math.Ceil(TestDuration/1000.0))+per >= taken)
}

func TestTenPerSecond(t *testing.T) {
	per := 10
	res, err := NewReservoir(uint64(per))
	assert.NoError(t, err)
	taken := takeOverTime(res, TestDuration)
	assert.True(t, int(math.Ceil(float64(TestDuration*per)/1000.0)) <= taken)
	assert.True(t, int(math.Ceil(float64(TestDuration*per)/1000.0))+per >= taken)
}

func TestDesiredRateTooLarge(t *testing.T) {
	per := 1e9
	_, err := NewReservoir(uint64(per))
	assert.EqualError(t, err, fmt.Sprintf("desired sampling capacity of %d is greater than maximum supported rate %d", uint64(per), uint64(1e8)))
}

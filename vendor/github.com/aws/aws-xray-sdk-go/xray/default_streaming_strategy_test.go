// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultStreamingStrategyMaxSegmentSize(t *testing.T) {
	dss, _ := NewDefaultStreamingStrategy()
	assert.Equal(t, dss.MaxSubsegmentCount, defaultMaxSubsegmentCount)
}

func TestDefaultStreamingStrategyMaxSegmentSizeParameterValidation(t *testing.T) {
	dss, e := NewDefaultStreamingStrategyWithMaxSubsegmentCount(-1)

	assert.Nil(t, dss)
	assert.Error(t, e, "maxSubsegmentCount must be a non-negative integer")
}

// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestNoNeedStreamingStrategy(t *testing.T) {
	seg := &Segment{}
	subSeg := &Segment{}
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &seg))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &subSeg))
	subSeg.ParentSegment = seg
	subSeg.parent = seg
	seg.ParentSegment = seg
	seg.Sampled = true
	seg.totalSubSegments = 1
	seg.rawSubsegments = append(seg.rawSubsegments, subSeg)
	assert.Equal(t, 1, len(packSegments(seg, nil)))
}

func TestStreamingSegmentsOnChildNode(t *testing.T) {
	seg := &Segment{}
	subSeg := &Segment{}
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &seg))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &subSeg))
	subSeg.parent = seg
	subSeg.ParentSegment = seg
	seg.Sampled = true
	seg.ParentSegment = seg
	seg.totalSubSegments = 22

	for i := 0; i < 22; i++ {
		seg.rawSubsegments = append(seg.rawSubsegments, subSeg)
	}

	out := packSegments(seg, nil)
	s := &Segment{}
	json.Unmarshal(out[2], s)
	assert.Equal(t, 20, len(s.Subsegments))
	assert.Equal(t, 3, len(out))
}

func TestStreamingSegmentsOnGrandchildNode(t *testing.T) {
	root := &Segment{}
	a := &Segment{}
	b := &Segment{}
	c := &Segment{}
	d := &Segment{}
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &root))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &a))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &b))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &c))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &d))

	root.ParentSegment = root
	root.Sampled = true
	a.ParentSegment = root
	b.ParentSegment = root
	c.ParentSegment = root
	d.ParentSegment = root
	a.parent = root
	b.parent = root
	c.parent = a
	d.parent = b
	root.totalSubSegments = 42
	root.rawSubsegments = append(root.rawSubsegments, a)
	root.rawSubsegments = append(root.rawSubsegments, b)

	for i := 0; i < 20; i++ {
		a.rawSubsegments = append(a.rawSubsegments, c)
	}
	for i := 0; i < 20; i++ {
		b.rawSubsegments = append(b.rawSubsegments, d)
	}
	assert.Equal(t, 23, len(packSegments(root, nil)))
}

func TestStreamingSegmentsTreeHasOnlyOneBranch(t *testing.T) {
	dss, _ := NewDefaultStreamingStrategyWithMaxSubsegmentCount(1)
	Configure(Config{StreamingStrategy: dss})
	segOne := &Segment{}
	segTwo := &Segment{}
	segThree := &Segment{}
	segFour := &Segment{}
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &segOne))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &segTwo))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &segThree))
	assert.NoError(t, json.Unmarshal([]byte(getTestSegment()), &segFour))

	segOne.ParentSegment = segOne
	segOne.Sampled = true
	segTwo.ParentSegment = segOne
	segTwo.parent = segOne
	segThree.ParentSegment = segOne
	segThree.parent = segTwo
	segFour.ParentSegment = segOne
	segFour.parent = segThree

	segOne.totalSubSegments = 3
	segOne.rawSubsegments = append(segOne.rawSubsegments, segTwo)
	segTwo.rawSubsegments = append(segTwo.rawSubsegments, segThree)
	segThree.rawSubsegments = append(segThree.rawSubsegments, segFour)

	assert.Equal(t, 3, len(packSegments(segOne, nil)))
	ResetConfig()
}

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "0123456789abcdef"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func getTestSegment() string {
	t := time.Now().Unix()
	hextime := fmt.Sprintf("%X", t)
	traceID := "1-" + hextime + "-" + randomString(24)
	message := fmt.Sprintf("{\"trace_id\": \"%s\", \"id\": \"%s\", \"start_time\": 1461096053.37518, "+
		"\"end_time\": 1461096053.4042, "+
		"\"name\": \"hello-1.mbfzqxzcpe.us-east-1.elasticbeanstalk.com\"}",
		traceID,
		randomString(16))
	return message
}

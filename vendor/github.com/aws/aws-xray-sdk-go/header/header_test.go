// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const ExampleTraceID string = "0-57ff426a-80c11c39b0c928905eb0828d"

func TestSampledEqualsOneFromString(t *testing.T) {
	h := FromString("Sampled=1")

	assert.Equal(t, Sampled, h.SamplingDecision)
	assert.Empty(t, h.TraceID)
	assert.Empty(t, h.ParentID)
	assert.Empty(t, h.AdditionalData)
}

func TestLonghFromString(t *testing.T) {
	h := FromString("Sampled=?;Root=" + ExampleTraceID + ";Parent=foo;Self=2;Foo=bar")

	assert.Equal(t, Requested, h.SamplingDecision)
	assert.Equal(t, ExampleTraceID, h.TraceID)
	assert.Equal(t, "foo", h.ParentID)
	assert.Equal(t, 1, len(h.AdditionalData))
	assert.Equal(t, "bar", h.AdditionalData["Foo"])
}

func TestLonghFromStringWithSpaces(t *testing.T) {
	h := FromString("Sampled=?; Root=" + ExampleTraceID + "; Parent=foo; Self=2; Foo=bar")

	assert.Equal(t, Requested, h.SamplingDecision)
	assert.Equal(t, ExampleTraceID, h.TraceID)
	assert.Equal(t, "foo", h.ParentID)
	assert.Equal(t, 1, len(h.AdditionalData))
	assert.Equal(t, "bar", h.AdditionalData["Foo"])
}

func TestSampledUnknownToString(t *testing.T) {
	h := &Header{}
	h.SamplingDecision = Unknown
	assert.Equal(t, "", h.String())
}

func TestSampledEqualsOneToString(t *testing.T) {
	h := &Header{}
	h.SamplingDecision = Sampled
	assert.Equal(t, "Sampled=1", h.String())
}

func TestSampledEqualsOneAndParentToString(t *testing.T) {
	h := &Header{}
	h.SamplingDecision = Sampled
	h.ParentID = "foo"
	assert.Equal(t, "Parent=foo;Sampled=1", h.String())
}

func TestLonghToString(t *testing.T) {
	h := &Header{}
	h.SamplingDecision = Sampled
	h.TraceID = ExampleTraceID
	h.ParentID = "foo"
	h.AdditionalData = make(map[string]string)
	h.AdditionalData["Foo"] = "bar"
	assert.Equal(t, "Root="+ExampleTraceID+";Parent=foo;Sampled=1;Foo=bar", h.String())
}

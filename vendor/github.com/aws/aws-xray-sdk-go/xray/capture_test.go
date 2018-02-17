// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-xray-sdk-go/strategy/exception"
	"github.com/stretchr/testify/assert"
)

func TestSimpleCapture(t *testing.T) {
	ctx, root := BeginSegment(context.Background(), "Test")
	err := Capture(ctx, "TestService", func(ctx1 context.Context) error {
		ctx = ctx1
		defer root.Close(nil)
		return nil
	})
	assert.NoError(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	assert.Equal(t, "Test", s.Name)
	assert.Equal(t, root.TraceID, s.TraceID)
	assert.Equal(t, root.ID, s.ID)
	assert.Equal(t, root.StartTime, s.StartTime)
	assert.Equal(t, root.EndTime, s.EndTime)
	assert.NotNil(t, s.Subsegments)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, "TestService", subseg.Name)
}

func TestCaptureAysnc(t *testing.T) {
	ctx, root := BeginSegment(context.Background(), "Test")
	CaptureAsync(ctx, "TestService", func(ctx1 context.Context) error {
		ctx = ctx1
		return nil
	})
	root.Close(nil)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	assert.Equal(t, "Test", s.Name)
	assert.Equal(t, root.TraceID, s.TraceID)
	assert.Equal(t, root.ID, s.ID)
	assert.Equal(t, root.StartTime, s.StartTime)
	assert.Equal(t, root.EndTime, s.EndTime)
	assert.NotNil(t, s.Subsegments)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, "TestService", subseg.Name)
}

func TestErrorCapture(t *testing.T) {
	ctx, root := BeginSegment(context.Background(), "Test")
	defaultStrategy, _ := exception.NewDefaultFormattingStrategy()
	err := Capture(ctx, "ErrorService", func(ctx1 context.Context) error {
		defer root.Close(nil)
		return defaultStrategy.Error("MyError")
	})

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, err.Error(), subseg.Cause.Exceptions[0].Message)
	assert.Equal(t, true, subseg.Fault)
	assert.Equal(t, "error", subseg.Cause.Exceptions[0].Type)
	assert.Equal(t, "TestErrorCapture.func1", subseg.Cause.Exceptions[0].Stack[0].Label)
	assert.Equal(t, "Capture", subseg.Cause.Exceptions[0].Stack[1].Label)
}

func TestPanicCapture(t *testing.T) {
	ctx, root := BeginSegment(context.Background(), "Test")
	var err error
	func() {
		defer func() {
			if p := recover(); p != nil {
				err = errors.New(p.(string))
			}
			root.Close(err)
		}()
		Capture(ctx, "PanicService", func(ctx1 context.Context) error {
			panic("MyPanic")
		})
	}()

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, err.Error(), subseg.Cause.Exceptions[0].Message)
	assert.Equal(t, "panic", subseg.Cause.Exceptions[0].Type)
	assert.Equal(t, "TestPanicCapture.func1.2", subseg.Cause.Exceptions[0].Stack[0].Label)
	assert.Equal(t, "Capture", subseg.Cause.Exceptions[0].Stack[1].Label)
	assert.Equal(t, "TestPanicCapture.func1", subseg.Cause.Exceptions[0].Stack[2].Label)
	assert.Equal(t, "TestPanicCapture", subseg.Cause.Exceptions[0].Stack[3].Label)
}

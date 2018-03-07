// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package exception

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiErrorReturnStringFormat(t *testing.T) {
	var err MultiError
	err = append(err, errors.New("error one"))
	err = append(err, errors.New("error two"))
	assert.Equal(t, "2 errors occurred:\n* error one\n* error two\n", err.Error())
}

func TestDefaultFormattingStrategyWithInvalidFrameCount(t *testing.T) {
	dss, e := NewDefaultFormattingStrategyWithDefinedErrorFrameCount(-1)
	ds, err := NewDefaultFormattingStrategyWithDefinedErrorFrameCount(33)
	assert.Nil(t, dss)
	assert.Nil(t, ds)
	assert.Error(t, e, "frameCount must be a non-negative integer and less than 32")
	assert.Error(t, err, "frameCount must be a non-negative integer and less than 32")
}

func TestNewDefaultFormattingStrategyWithValidFrameCount(t *testing.T) {
	dss, e := NewDefaultFormattingStrategyWithDefinedErrorFrameCount(10)
	assert.Nil(t, e)
	assert.Equal(t, 10, dss.FrameCount)
}

func TestError(t *testing.T) {
	defs, _ := NewDefaultFormattingStrategy()

	err := defs.Error("Test")
	stack := convertStack(err.StackTrace())

	assert.Equal(t, "Test", err.Error())
	assert.Equal(t, "error", err.Type)
	assert.Equal(t, "TestError", stack[0].Label)
}

func TestErrorf(t *testing.T) {
	defs, _ := NewDefaultFormattingStrategy()

	err := defs.Errorf("Test")
	stack := convertStack(err.StackTrace())

	assert.Equal(t, "Test", err.Error())
	assert.Equal(t, "error", err.Type)
	assert.Equal(t, "TestErrorf", stack[0].Label)
}

func TestPanic(t *testing.T) {
	defs, _ := NewDefaultFormattingStrategy()

	var err *XRayError
	func() {
		defer func() {
			err = defs.Panic(recover().(string))
		}()
		panic("Test")
	}()
	stack := convertStack(err.StackTrace())

	assert.Equal(t, "Test", err.Error())
	assert.Equal(t, "panic", err.Type)
	assert.Equal(t, "TestPanic.func1", stack[0].Label)
	assert.Equal(t, "TestPanic", stack[1].Label)
}

func TestPanicf(t *testing.T) {
	defs, _ := NewDefaultFormattingStrategy()

	var err *XRayError
	func() {
		defer func() {
			err = defs.Panicf("%v", recover())
		}()
		panic("Test")
	}()
	stack := convertStack(err.StackTrace())

	assert.Equal(t, "Test", err.Error())
	assert.Equal(t, "panic", err.Type)
	assert.Equal(t, "TestPanicf.func1", stack[0].Label)
	assert.Equal(t, "TestPanicf", stack[1].Label)
}

func TestExceptionFromError(t *testing.T) {
	defaultStrategy := &DefaultFormattingStrategy{}

	err := defaultStrategy.ExceptionFromError(errors.New("new error"))

	assert.NotNil(t, err.ID)
	assert.Equal(t, "new error", err.Message)
	assert.Equal(t, "error", err.Type)
}

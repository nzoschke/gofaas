// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-xray-sdk-go/strategy/ctxmissing"
	"github.com/aws/aws-xray-sdk-go/strategy/exception"
	"github.com/aws/aws-xray-sdk-go/strategy/sampling"
	"github.com/stretchr/testify/assert"
)

type TestSamplingStrategy struct{}

type TestExceptionFormattingStrategy struct{}

type TestStreamingStrategy struct{}

type TestContextMissingStrategy struct{}

func (tss *TestSamplingStrategy) ShouldTrace(serviceName string, path string, method string) bool {
	return true
}

func (tefs *TestExceptionFormattingStrategy) Error(message string) *exception.XRayError {
	return &exception.XRayError{}
}

func (tefs *TestExceptionFormattingStrategy) Errorf(message string, args ...interface{}) *exception.XRayError {
	return &exception.XRayError{}
}

func (tefs *TestExceptionFormattingStrategy) Panic(message string) *exception.XRayError {
	return &exception.XRayError{}
}

func (tefs *TestExceptionFormattingStrategy) Panicf(message string, args ...interface{}) *exception.XRayError {
	return &exception.XRayError{}
}

func (tefs *TestExceptionFormattingStrategy) ExceptionFromError(err error) exception.Exception {
	return exception.Exception{}
}

func (sms *TestStreamingStrategy) RequiresStreaming(seg *Segment) bool {
	return false
}

func (sms *TestStreamingStrategy) StreamCompletedSubsegments(seg *Segment) [][]byte {
	var test [][]byte
	return test
}

func (cms *TestContextMissingStrategy) ContextMissing(v interface{}) {
	fmt.Sprintf("Test ContextMissing Strategy %v", v)
}

func stashEnv() []string {
	env := os.Environ()
	os.Clearenv()
	return env
}

func popEnv(env []string) {
	os.Clearenv()
	for _, e := range env {
		p := strings.SplitN(e, "=", 2)
		os.Setenv(p[0], p[1])
	}
}

func ResetConfig() {
	ss, _ := sampling.NewLocalizedStrategy()
	efs, _ := exception.NewDefaultFormattingStrategy()
	sms, _ := NewDefaultStreamingStrategy()
	cms := ctxmissing.NewDefaultRuntimeErrorStrategy()

	Configure(Config{
		DaemonAddr:                  "127.0.0.1:2000",
		LogLevel:                    "info",
		LogFormat:                   "%Date(2006-01-02T15:04:05Z07:00) [%Level] %Msg%n",
		SamplingStrategy:            ss,
		StreamingStrategy:           sms,
		ExceptionFormattingStrategy: efs,
		ContextMissingStrategy:      cms,
	})
}

func TestDefaultConfigureParameters(t *testing.T) {
	daemonAddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 2000}
	logLevel := "info"
	logFormat := "%Date(2006-01-02T15:04:05Z07:00) [%Level] %Msg%n"
	ss, _ := sampling.NewLocalizedStrategy()
	efs, _ := exception.NewDefaultFormattingStrategy()
	sms, _ := NewDefaultStreamingStrategy()
	cms := ctxmissing.NewDefaultRuntimeErrorStrategy()

	assert.Equal(t, daemonAddr, privateCfg.daemonAddr)
	assert.Equal(t, logLevel, privateCfg.logLevel.String())
	assert.Equal(t, logFormat, privateCfg.logFormat)
	assert.Equal(t, ss, privateCfg.samplingStrategy)
	assert.Equal(t, efs, privateCfg.exceptionFormattingStrategy)
	assert.Equal(t, "", privateCfg.serviceVersion)
	assert.Equal(t, sms, privateCfg.streamingStrategy)
	assert.Equal(t, cms, privateCfg.contextMissingStrategy)
}

func TestSetConfigureParameters(t *testing.T) {
	daemonAddr := "127.0.0.1:3000"
	logLevel := "error"
	logFormat := "[%Level] %Msg%n"
	serviceVersion := "TestVersion"

	ss := &TestSamplingStrategy{}
	efs := &TestExceptionFormattingStrategy{}
	sms := &TestStreamingStrategy{}
	cms := &TestContextMissingStrategy{}

	Configure(Config{
		DaemonAddr:                  daemonAddr,
		ServiceVersion:              serviceVersion,
		SamplingStrategy:            ss,
		ExceptionFormattingStrategy: efs,
		StreamingStrategy:           sms,
		ContextMissingStrategy:      cms,
		LogLevel:                    logLevel,
		LogFormat:                   logFormat,
	})

	assert.Equal(t, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3000}, privateCfg.daemonAddr)
	assert.Equal(t, logLevel, privateCfg.logLevel.String())
	assert.Equal(t, logFormat, privateCfg.logFormat)
	assert.Equal(t, ss, privateCfg.samplingStrategy)
	assert.Equal(t, efs, privateCfg.exceptionFormattingStrategy)
	assert.Equal(t, sms, privateCfg.streamingStrategy)
	assert.Equal(t, cms, privateCfg.contextMissingStrategy)
	assert.Equal(t, serviceVersion, privateCfg.serviceVersion)

	ResetConfig()
}

func TestSetDaemonAddressEnvironmentVariable(t *testing.T) {
	env := stashEnv()
	defer popEnv(env)
	daemonAddr := "127.0.0.1:3000"
	os.Setenv("AWS_XRAY_DAEMON_ADDRESS", "127.0.0.1:4000")
	Configure(Config{DaemonAddr: daemonAddr})
	assert.Equal(t, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4000}, privateCfg.daemonAddr)
	os.Unsetenv("AWS_XRAY_DAEMON_ADDRESS")

	ResetConfig()
}

func TestSetContextMissingEnvironmentVariable(t *testing.T) {
	env := stashEnv()
	defer popEnv(env)
	cms := ctxmissing.NewDefaultLogErrorStrategy()
	r := ctxmissing.NewDefaultRuntimeErrorStrategy()
	os.Setenv("AWS_XRAY_CONTEXT_MISSING", "RUNTIME_ERROR")
	Configure(Config{ContextMissingStrategy: cms})
	assert.Equal(t, r, privateCfg.contextMissingStrategy)
	os.Unsetenv("AWS_XRAY_CONTEXT_MISSING")

	ResetConfig()
}

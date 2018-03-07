// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		b := []byte(`200 - OK`)
		w.Write(b)
	})

	ts := httptest.NewServer(Handler(NewFixedSegmentNamer("test"), handler))

	req := httptest.NewRequest("POST", ts.URL, strings.NewReader(""))
	req.Header.Set("User-Agent", "UnitTest")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")

	_, err := http.DefaultTransport.RoundTrip(req)
	assert.NoError(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)

	assert.Equal(t, http.StatusOK, s.HTTP.Response.Status)
	assert.Equal(t, "POST", s.HTTP.Request.Method)
	assert.Equal(t, ts.URL+"/", s.HTTP.Request.URL)
	assert.Equal(t, "127.0.0.1", s.HTTP.Request.ClientIP)
	assert.Equal(t, "UnitTest", s.HTTP.Request.UserAgent)
}

func TestNonRootHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(Handler(NewFixedSegmentNamer("test"), handler))

	req := httptest.NewRequest("DELETE", ts.URL, strings.NewReader(""))
	req.Header.Set("x-amzn-trace-id", "Root=fakeid; Parent=reqid; Sampled=1")

	_, err := http.DefaultTransport.RoundTrip(req)
	assert.NoError(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)

	assert.Equal(t, "fakeid", s.TraceID)
	assert.Equal(t, "reqid", s.ParentID)
}

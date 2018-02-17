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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rt *roundtripper

func init() {
	rt = &roundtripper{
		Base: http.DefaultTransport,
	}
}

func TestNilClient(t *testing.T) {
	c := Client(nil)
	assert.Equal(t, http.DefaultClient.Jar, c.Jar)
	assert.Equal(t, http.DefaultClient.Timeout, c.Timeout)
	assert.Equal(t, &roundtripper{Base: http.DefaultTransport}, c.Transport)
}

func TestRoundTripper(t *testing.T) {
	ht := http.DefaultTransport
	rt := RoundTripper(ht)
	assert.Equal(t, &roundtripper{Base: http.DefaultTransport}, rt)
}

func TestRoundTrip(t *testing.T) {
	var responseContentLength int
	var headers XRayHeaders
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = ParseHeadersForTest(r.Header)
		b := []byte(`200 - Nothing to see`)
		responseContentLength = len(b)
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))

	defer ts.Close()

	reader := strings.NewReader("")
	ctx, root := BeginSegment(context.Background(), "Test")
	req := httptest.NewRequest("GET", ts.URL, reader)
	req = req.WithContext(ctx)
	_, err := rt.RoundTrip(req)
	root.Close(nil)
	assert.NoError(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, "remote", subseg.Namespace)
	assert.Equal(t, "GET", subseg.HTTP.Request.Method)
	assert.Equal(t, ts.URL, subseg.HTTP.Request.URL)
	assert.Equal(t, 200, subseg.HTTP.Response.Status)
	assert.Equal(t, responseContentLength, subseg.HTTP.Response.ContentLength)
	assert.Equal(t, headers.RootTraceID, s.TraceID)
}

func TestRoundTripWithError(t *testing.T) {
	var responseContentLength int
	var headers XRayHeaders
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = ParseHeadersForTest(r.Header)
		b := []byte(`403 - Nothing to see`)
		responseContentLength = len(b)
		w.WriteHeader(http.StatusForbidden)
		w.Write(b)
	}))

	defer ts.Close()

	reader := strings.NewReader("")
	ctx, root := BeginSegment(context.Background(), "Test")
	req := httptest.NewRequest("GET", ts.URL, reader)
	req = req.WithContext(ctx)
	_, err := rt.RoundTrip(req)
	root.Close(nil)
	assert.NoError(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, "remote", subseg.Namespace)
	assert.Equal(t, "GET", subseg.HTTP.Request.Method)
	assert.Equal(t, ts.URL, subseg.HTTP.Request.URL)
	assert.Equal(t, 403, subseg.HTTP.Response.Status)
	assert.Equal(t, responseContentLength, subseg.HTTP.Response.ContentLength)
	assert.Equal(t, headers.RootTraceID, s.TraceID)
}

func TestRoundTripWithThrottle(t *testing.T) {
	var responseContentLength int
	var headers XRayHeaders
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = ParseHeadersForTest(r.Header)

		b := []byte(`429 - Nothing to see`)
		responseContentLength = len(b)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write(b)
	}))

	defer ts.Close()

	reader := strings.NewReader("")
	ctx, root := BeginSegment(context.Background(), "Test")
	req := httptest.NewRequest("GET", ts.URL, reader)
	req = req.WithContext(ctx)
	_, err := rt.RoundTrip(req)
	root.Close(nil)
	assert.NoError(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, "remote", subseg.Namespace)
	assert.Equal(t, "GET", subseg.HTTP.Request.Method)
	assert.Equal(t, ts.URL, subseg.HTTP.Request.URL)
	assert.Equal(t, 429, subseg.HTTP.Response.Status)
	assert.Equal(t, responseContentLength, subseg.HTTP.Response.ContentLength)
	assert.Equal(t, headers.RootTraceID, s.TraceID)
}

func TestRoundTripFault(t *testing.T) {
	var responseContentLength int
	var headers XRayHeaders
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = ParseHeadersForTest(r.Header)

		b := []byte(`510 - Nothing to see`)
		responseContentLength = len(b)
		w.WriteHeader(http.StatusNotExtended)
		w.Write(b)
	}))

	defer ts.Close()

	reader := strings.NewReader("")
	ctx, root := BeginSegment(context.Background(), "Test")
	req := httptest.NewRequest("GET", ts.URL, reader)
	req = req.WithContext(ctx)
	_, err := rt.RoundTrip(req)
	root.Close(nil)
	assert.NoError(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, "remote", subseg.Namespace)
	assert.Equal(t, "GET", subseg.HTTP.Request.Method)
	assert.Equal(t, ts.URL, subseg.HTTP.Request.URL)
	assert.Equal(t, 510, subseg.HTTP.Response.Status)
	assert.Equal(t, responseContentLength, subseg.HTTP.Response.ContentLength)
	assert.Equal(t, headers.RootTraceID, s.TraceID)
}

func TestBadRoundTrip(t *testing.T) {
	ctx, root := BeginSegment(context.Background(), "Test")
	reader := strings.NewReader("")
	req := httptest.NewRequest("GET", "httpz://localhost:8000", reader)
	req = req.WithContext(ctx)
	_, err := rt.RoundTrip(req)
	root.Close(nil)
	assert.Error(t, err)

	s, e := TestDaemon.Recv()
	assert.NoError(t, e)
	subseg := &Segment{}
	assert.NoError(t, json.Unmarshal(s.Subsegments[0], &subseg))
	assert.Equal(t, fmt.Sprintf("%v", err), subseg.Cause.Exceptions[0].Message)
}

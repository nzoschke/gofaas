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
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	listenerAddr = &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 2000,
	}

	TestDaemon = &Testdaemon{
		Channel: make(chan *result, 200),
	}
)

func init() {
	if TestDaemon.Connection == nil {
		conn, err := net.ListenUDP("udp", listenerAddr)
		if err != nil {
			panic(err)
		}

		TestDaemon.Connection = conn
		go TestDaemon.Run()
	}
}

type Testdaemon struct {
	Connection *net.UDPConn
	Channel    chan *result
	Done       bool
}
type result struct {
	Segment *Segment
	Error   error
}

func (td *Testdaemon) Run() {
	buffer := make([]byte, 64000)
	for !td.Done {
		n, _, err := td.Connection.ReadFromUDP(buffer)
		if err != nil {
			td.Channel <- &result{nil, err}
			continue
		}

		buffered := buffer[len(Header):n]

		seg := &Segment{}
		err = json.Unmarshal(buffered, &seg)
		if err != nil {
			td.Channel <- &result{nil, err}
			continue
		}

		seg.Sampled = true
		td.Channel <- &result{seg, err}
	}
}

func (td *Testdaemon) Recv() (*Segment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	select {
	case r := <-td.Channel:
		return r.Segment, r.Error
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type XRayHeaders struct {
	RootTraceID string
	ParentID    string
	Sampled     bool
}

func ParseHeadersForTest(h http.Header) XRayHeaders {
	m := parseHeaders(h)
	s, _ := strconv.ParseBool(m["Sampled"])

	return XRayHeaders{
		RootTraceID: m["Root"],
		ParentID:    m["Parent"],
		Sampled:     s,
	}
}

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v2

import (
	"io"
	"net"
	"sync"
)

// AsyncClient asynchronously publishes events to lumberjack endpoint. On ACK a
// provided callback function will be called. The number of in-flight publish
// requests is configurable but limited. Once the limit has been reached, the
// client will block publish requests until the lumberjack server did ACK some
// queued publish requests.
type AsyncClient struct {
	cl *Client

	inflight int
	ch       chan ackMessage
	wg       sync.WaitGroup
}

type ackMessage struct {
	cb      AsyncSendCallback
	seq     uint32
	dropped int
	err     error
}

// AsyncSendCallback callback function. Upon completion seq contains the last
// ACKed event's index. The count starts with 1. The err argument contains the latest
// error encountered by lumberjack client.
//
// Note: The callback MUST not block. In case callback is trying to republish
// not ACKed events, care must be taken not to deadlock the AsyncClient when calling
// Send.
type AsyncSendCallback func(seq uint32, err error)

// NewAsyncClientWith creates a new AsyncClient from low-level lumberjack v2 Client.
// The inflight argument sets number of active publish requests.
func NewAsyncClientWith(cl *Client, inflight int) (*AsyncClient, error) {
	c := &AsyncClient{
		cl:       cl,
		inflight: inflight,
	}

	c.startACK()
	return c, nil
}

// NewAsyncClientWithConn creates a new AsyncClient from an active connection.
func NewAsyncClientWithConn(c net.Conn, inflight int, opts ...Option) (*AsyncClient, error) {
	cl, err := NewWithConn(c, opts...)
	if err != nil {
		return nil, err
	}
	return NewAsyncClientWith(cl, inflight)
}

// AsyncDial connects to lumberjack server and returns new AsyncClient. On error
// no AsyncClient is being created.
func AsyncDial(address string, inflight int, opts ...Option) (*AsyncClient, error) {
	cl, err := Dial(address, opts...)
	if err != nil {
		return nil, err
	}
	return NewAsyncClientWith(cl, inflight)
}

// AsyncDialWith uses provided dialer to connect to lumberjack server. On error
// no AsyncClient is being returned.
func AsyncDialWith(
	dial func(network, address string) (net.Conn, error),
	address string,
	inflight int,
	opts ...Option,
) (*AsyncClient, error) {
	cl, err := DialWith(dial, address, opts...)
	if err != nil {
		return nil, err
	}
	return NewAsyncClientWith(cl, inflight)
}

// Close closes the client, so no new events can be published anymore. The
// underlying network connection will be closed too. Returns an error if
// underlying net.Conn errors on Close.
//
// All inflight requests will be cancelled, returning EOF if no other error has
// been encountered due to underlying network connection being closed.
//
// The client gives no guarantees regarding published events. There is a chance
// events will be processed by server, even though connection has been closed.
func (c *AsyncClient) Close() error {
	err := c.cl.Close()
	c.stopACK()
	return err
}

// Send publishes a new batch of events by JSON-encoding given batch.
// Send blocks if maximum number of allowed asynchrounous calls is still active.
// Upon completion cb will be called with last ACKed index into active batch.
// Returns error if communication or serialization to JSON failed.
func (c *AsyncClient) Send(cb AsyncSendCallback, data []interface{}) error {
	dropped, err := c.cl.Send(data)
	if err != nil || dropped == len(data) {
		c.ch <- ackMessage{
			seq:     0,
			dropped: dropped,
			cb:      cb,
			err:     err,
		}
		return err
	}

	c.ch <- ackMessage{
		seq:     uint32(len(data) - dropped),
		dropped: dropped,
		cb:      cb,
		err:     nil,
	}
	return nil
}

func (c *AsyncClient) startACK() {
	c.ch = make(chan ackMessage, c.inflight)
	c.wg.Add(1)
	go c.ackLoop()
}

func (c *AsyncClient) stopACK() {
	close(c.ch)
	c.wg.Wait()
}

func (c *AsyncClient) ackLoop() {
	var seq uint32
	var err error

	// drain ack queue on error/exit
	defer func() {
		if err == nil {
			err = io.EOF
		}
		for msg := range c.ch {
			if msg.err != nil {
				err = msg.err
			}
			msg.cb(0, err)
		}
	}()
	defer c.wg.Done()

	for msg := range c.ch {
		if msg.err != nil || msg.seq == 0 {
			err = msg.err
			msg.cb(msg.seq+uint32(msg.dropped), msg.err)
			return
		}

		seq, err = c.cl.AwaitACK(msg.seq)
		msg.cb(seq+uint32(msg.dropped), err)
		if err != nil {
			c.cl.Close()
			return
		}
	}
}

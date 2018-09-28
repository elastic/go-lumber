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

package v1

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/klauspost/compress/zlib"

	"github.com/elastic/go-lumber/lj"
	"github.com/elastic/go-lumber/log"
	protocol "github.com/elastic/go-lumber/protocol/v1"
)

type reader struct {
	in      *bufio.Reader
	conn    net.Conn
	timeout time.Duration
	buf     []byte
}

func newReader(c net.Conn, to time.Duration) *reader {
	r := &reader{
		in:      bufio.NewReader(c),
		conn:    c,
		timeout: to,
		buf:     make([]byte, 0, 64),
	}
	return r
}

func (r *reader) ReadBatch() (*lj.Batch, error) {
	// 1. read window size
	var win [6]byte
	_ = r.conn.SetReadDeadline(time.Time{}) // wait for next batch without timeout
	if err := readFull(r.in, win[:]); err != nil {
		return nil, err
	}

	if win[0] != protocol.CodeVersion && win[1] != protocol.CodeWindowSize {
		log.Printf("Expected window from. Received %v", win[0:1])
		return nil, ErrProtocolError
	}

	count := int(binary.BigEndian.Uint32(win[2:]))
	if count == 0 {
		return nil, nil
	}

	if err := r.conn.SetReadDeadline(time.Now().Add(r.timeout)); err != nil {
		return nil, err
	}

	events, seq, err := r.readEvents(r.in, make([]interface{}, 0, count))
	if events == nil || err != nil {
		log.Printf("readEvents failed with: %v", err)
		return nil, err
	}

	return lj.NewBatchWithSequence(events, seq), nil
}

func (r *reader) readEvents(in io.Reader, events []interface{}) ([]interface{}, uint32, error) {
	var seq uint32 = 0
	for len(events) < cap(events) {
		var hdr [2]byte
		if err := readFull(in, hdr[:]); err != nil {
			return nil, 0, err
		}

		if hdr[0] != protocol.CodeVersion {
			log.Println("Event protocol version error")
			return nil, 0, ErrProtocolError
		}

		switch hdr[1] {
		case protocol.CodeDataFrame:
			event, tmpseq, err := r.readEvent(in)
			if err != nil {
				log.Printf("failed to read json event with: %v\n", err)
				return nil, 0, err
			}
			if tmpseq > seq {
				seq = tmpseq
			}
			events = append(events, event)
		case protocol.CodeCompressed:
			readEvents, tmpseq, err := r.readCompressed(in, events)
			if err != nil {
				return nil, 0, err
			}
			if tmpseq > seq {
				seq = tmpseq
			}
			events = readEvents
		default:
			log.Printf("Unknown frame type: %v", hdr[1])
			return nil, 0, ErrProtocolError
		}
	}
	return events, seq, nil
}

func (r *reader) readCompressed(in io.Reader, events []interface{}) ([]interface{}, uint32, error) {
	var seq uint32
	var hdr [4]byte
	if err := readFull(in, hdr[:]); err != nil {
		return nil, 0, err
	}

	payloadSz := binary.BigEndian.Uint32(hdr[:])
	limit := io.LimitReader(in, int64(payloadSz))
	reader, err := zlib.NewReader(limit)
	if err != nil {
		log.Printf("Failed to initialized zlib reader %v\n", err)
		return nil, 0, err
	}

	events, seq, err = r.readEvents(reader, events)
	if err != nil {
		_ = reader.Close()
		return nil, 0, err
	}
	if err := reader.Close(); err != nil {
		return nil, 0, err
	}

	// consume final bytes from limit reader
	for {
		var tmp [16]byte
		if _, err := limit.Read(tmp[:]); err != nil {
			if err != io.EOF {
				return nil, 0, err
			}
			break
		}
	}
	return events, seq, nil
}

func (r *reader) readEvent(in io.Reader) (interface{}, uint32, error) {
	var hdr [8]byte
	if err := readFull(in, hdr[:]); err != nil {
		return nil, 0, err
	}

	readString := func() (string, error) {
		var bufBytes [4]byte
		if err := readFull(in, bufBytes[:]); err != nil {
			return "", err
		}

		bytes := int(binary.BigEndian.Uint32(bufBytes[:]))
		if bytes > len(r.buf) {
			r.buf = make([]byte, bytes)
		}

		buf := r.buf[:bytes]
		if err := readFull(in, buf); err != nil {
			return "", err
		}

		return string(buf[:]), nil
	}

	seq := binary.BigEndian.Uint32(hdr[0:4])
	event := map[string]string{}
	pairs := int(binary.BigEndian.Uint32(hdr[4:]))
	for i := 0; i < pairs; i++ {
		k, err := readString()
		if err != nil {
			return nil, 0, err
		}

		v, err := readString()
		if err != nil {
			return nil, 0, err
		}

		event[k] = v
	}
	return event, seq, nil
}

func readFull(in io.Reader, buf []byte) error {
	_, err := io.ReadFull(in, buf)
	return err
}

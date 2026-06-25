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
	"encoding/binary"
	"net"
	"testing"
	"time"

	protocol "github.com/elastic/go-lumber/protocol/v1" //nolint:depguard // We are testing.
)

func TestReadBatchWindowTooLarge(t *testing.T) {
	client, srv := net.Pipe()
	defer client.Close()
	defer srv.Close()

	r := newReader(srv, 5*time.Second, 100)

	go func() {
		var frame [6]byte
		frame[0] = protocol.CodeVersion
		frame[1] = protocol.CodeWindowSize
		binary.BigEndian.PutUint32(frame[2:], 101)
		_, _ = client.Write(frame[:])
	}()

	_, err := r.ReadBatch()
	if err != ErrWindowTooLarge { //nolint:errorlint // Never wrapped.
		t.Fatalf("ReadBatch() error = %v, want %v", err, ErrWindowTooLarge)
	}
}

func TestReadBatchProtocolValidation(t *testing.T) {
	tests := []struct {
		name    string
		frame   [6]byte
		wantErr error
	}{
		{
			name:    "wrong version byte",
			frame:   [6]byte{'X', protocol.CodeWindowSize, 0, 0, 0, 1},
			wantErr: ErrProtocolError,
		},
		{
			name:    "wrong type byte",
			frame:   [6]byte{protocol.CodeVersion, 'X', 0, 0, 0, 1},
			wantErr: ErrProtocolError,
		},
		{
			name:    "both wrong",
			frame:   [6]byte{'X', 'Y', 0, 0, 0, 1},
			wantErr: ErrProtocolError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, srv := net.Pipe()
			defer client.Close()
			defer srv.Close()

			r := newReader(srv, 5*time.Second, 10000)

			go func() {
				_, _ = client.Write(test.frame[:])
			}()

			_, err := r.ReadBatch()
			if err != test.wantErr { //nolint:errorlint // Never wrapped.
				t.Fatalf("ReadBatch() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestReadBatchMaxWindowSizeDisabled(t *testing.T) {
	client, srv := net.Pipe()
	defer client.Close()
	defer srv.Close()

	r := newReader(srv, 5*time.Second, 0)

	go func() {
		var frame [6]byte
		frame[0] = protocol.CodeVersion
		frame[1] = protocol.CodeWindowSize
		binary.BigEndian.PutUint32(frame[2:], 50000)
		_, _ = client.Write(frame[:])
		client.Close()
	}()

	_, err := r.ReadBatch()
	if err == ErrWindowTooLarge { //nolint:errorlint // Never wrapped.
		t.Fatal("ReadBatch() returned ErrWindowTooLarge with maxWindowSize=0")
	}
}

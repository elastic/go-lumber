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

import "fmt"

// EventSerializationError indicates that the event
// serialization failed.
type EventSerializationError struct {
	At     int         // Batch index of the event that failed to serialize.
	Data   interface{} // Original structure that failed to serialize.
	Reason error       // Original error reported by the serializer.
}

var _ error = (*EventSerializationError)(nil)

func (e *EventSerializationError) Unwrap() error { return e.Reason }

func (e *EventSerializationError) Error() string {
	return fmt.Sprintf("failed to serialize event: %v", e.Reason)
}

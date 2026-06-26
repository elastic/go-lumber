# Lumberjack Protocol Version 2 Specification

## Abstract

This document specifies the Lumberjack Protocol Version 2 (LJ-v2), a binary protocol designed for efficient, reliable transmission of structured data (logs, metrics, events) from clients to servers. The protocol is used in Elastic's Beats framework and optimized for high throughput with reliability guarantees.

## 1. Introduction

Lumberjack v2 is a binary protocol providing reliable, efficient transport of structured data with these key features:

- Binary framing for parsing efficiency
- Sequence numbering and acknowledgments for reliability
- Batch processing for improved throughput
- Compression support for bandwidth optimization
- Window-based flow control

The protocol is designed to be simple to implement while providing robust delivery guarantees.

## 2. Transport Layer

- **Protocol**: TCP (optionally with TLS encryption)
- **Port**: Default 5044 (configurable)
- **Connection Model**: Long-lived, persistent connections
- **Encryption**: TLS strongly recommended but optional

## 3. Data Types and Encodings

| Type | Size | Description |
|------|------|-------------|
| byte | 1 byte | Unsigned 8-bit integer |
| uint32 | 4 bytes | Unsigned 32-bit integer, network byte order (big-endian) |
| string | variable | UTF-8 encoded string prefixed with uint32 length |
| json | variable | UTF-8 encoded JSON object prefixed with uint32 length |

All integers are encoded in network byte order (big-endian).

## 4. Protocol Frames

All protocol communications after connection establishment use a frame-based format.

### 4.1. Frame Header

Every frame begins with a 6-byte header:

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Version    |  Frame Type   |           Payload Size        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|        Payload Size (cont.)   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **Version**: 1 byte, Value '2' (ASCII 0x32) for Lumberjack v2
- **Frame Type**: 1 byte, indicates frame type (see section 4.2)
- **Payload Size**: 4 bytes (uint32), length of payload in bytes

### 4.2. Frame Types

| Type | ASCII Value | Hex Value | Direction | Description |
|------|-------------|-----------|-----------|-------------|
| Window Size | '1' | 0x31 | Client → Server | Sets window size |
| JSON Data | '2' | 0x32 | Client → Server | Uncompressed events |
| Compressed Data | '3' | 0x33 | Client → Server | Compressed events |
| Acknowledgment | 'A' | 0x41 | Server → Client | Acknowledges events |

## 5. Frame Specifications

### 5.1. Window Size Frame (Type '1')

Sent by the client to declare maximum number of unacknowledged events.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Version (2)  |  Type ('1')   |           Size (4)            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         Size (cont.)          |          Window Size          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Window Size (cont.)      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **Payload Size**: Always 4 bytes
- **Window Size**: uint32, maximum number of unacknowledged events

### 5.2. JSON Data Frame (Type '2')

Transmits uncompressed JSON events.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Version (2)  |  Type ('2')   |           Payload Size        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Payload Size (cont.)     |         Sequence Number       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     Sequence Number (cont.)   |          Event Count          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Event Count (cont.)      |      JSON Events payload...   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **Sequence Number**: uint32, monotonically increasing sequence identifier
- **Event Count**: uint32, number of JSON events in the payload
- **JSON Events**: Series of length-prefixed JSON objects

Each JSON event is encoded as:
```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         JSON Length                           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         JSON Content...                       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **JSON Length**: uint32, length of the JSON object in bytes
- **JSON Content**: UTF-8 encoded JSON object

### 5.3. Compressed Data Frame (Type '3')

Transmits zlib-compressed JSON events.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Version (2)  |  Type ('3')   |           Payload Size        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Payload Size (cont.)     |         Sequence Number       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     Sequence Number (cont.)   |          Event Count          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Event Count (cont.)      | Compressed JSON Events payload |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **Sequence Number**: uint32, monotonically increasing sequence identifier
- **Event Count**: uint32, number of JSON events in the compressed payload
- **Compressed JSON Events**: zlib-compressed series of length-prefixed JSON events

When decompressed, the format is identical to the JSON Data frame payload.

### 5.4. Acknowledgment Frame (Type 'A')

Sent by the server to acknowledge receipt of events.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Version (2)  |  Type ('A')   |           Size (4)            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         Size (cont.)          |         Sequence Number       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     Sequence Number (cont.)   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **Payload Size**: Always 4 bytes
- **Sequence Number**: uint32, acknowledges all events up to and including this sequence number

## 6. Protocol Operation

### 6.1. Connection Establishment

1. Client establishes TCP connection to server
2. TLS handshake (if enabled)
3. Client sends Window Size frame (required before sending events)

### 6.2. Data Flow

**Normal Operation:**
1. Client batches events up to configured batch size
2. Client assigns sequence number to batch
3. Client sends JSON Data or Compressed Data frame
4. Server processes events
5. Server sends Acknowledgment frame
6. Client releases acknowledged events from memory

**Flow Control:**
- Client must not send more unacknowledged events than the declared window size
- Client tracks sent events that haven't been acknowledged
- Server acknowledges event batches as they are processed

### 6.3. Connection Termination

Either party may close the connection at any time. Proper termination sequence:
1. Client finishes sending all pending events
2. Client waits for all acknowledgments
3. Client or server closes the TCP connection

## 7. Client Configuration Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `window_size` | Maximum unacknowledged events | 1024 |
| `compression_level` | zlib compression level (0-9, 0=disabled) | 3 |
| `compression_threshold` | Minimum batch size to trigger compression (bytes) | 1024 |
| `batch_size` | Maximum events per batch | 100 |
| `timeout` | Connection/acknowledgment timeout | 30s |
| `max_retries` | Maximum reconnection attempts | 3 |
| `backoff_init` | Initial backoff delay | 1s |
| `backoff_max` | Maximum backoff delay | 30s |

## 8. Server Configuration Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `max_window_size` | Maximum allowed client window size | 1024 |
| `max_batch_size` | Maximum allowed events per batch | 256 |
| `ack_timeout` | Maximum time before sending acknowledgment | 5s |
| `read_timeout` | Maximum time to wait for complete frame | 30s |
| `idle_timeout` | Disconnect idle clients after this period | 60s |

## 9. Error Handling

### 9.1. Connection Errors

Clients should implement:
- Reconnection with exponential backoff
- Tracking of pending events for retransmission
- Proper resource cleanup

### 9.2. Protocol Errors

When receiving invalid or unexpected data:
1. Log detailed error information
2. Close the connection
3. For clients: attempt reconnection with backoff
4. For servers: accept new connections

Common protocol errors:
- Invalid frame type
- Invalid payload size
- Decompression failures
- JSON parsing errors
- Sequence number errors

## 10. Implementation Guidance

### 10.1. Best Practices

1. **Buffering**: Clients should maintain persistent buffers for events
2. **Batching**: Group events into reasonably sized batches for efficiency
3. **Compression**: Use compression for large batches to save bandwidth
4. **Timeouts**: Set reasonable timeouts for all operations
5. **TLS**: Use TLS with proper certificate validation

### 10.2. Event Format Recommendations

While the protocol doesn't mandate specific JSON fields, [common practice includes](https://wikitech.wikimedia.org/wiki/Logstash/Common_Logging_Schema):

```json
{
  "@timestamp": "2023-05-20T12:00:00.000Z",
  "message": "Log message content",
  "log": {
    "file": {
      "path": "/var/log/syslog"
    },
    "offset": 12345
  },
  "host": {
    "name": "server1.example.com"
  },
  "event": {
    "dataset": "system.syslog"
  }
}
```

### 10.3. Language-Specific Considerations

#### PHP
- Use `pack()`/`unpack()` for binary encoding/decoding
- `gzcompress()`/`gzuncompress()` for zlib compression
- Socket functions or streams for TCP communication
- `json_encode()`/`json_decode()` with error checking

#### Rust
- Use `byteorder` crate for network byte order
- `flate2` crate for zlib compression
- `serde_json` for JSON handling
- Consider `tokio` for async I/O

#### TypeScript/JavaScript
- Use `Buffer` for binary operations
- `zlib` module for compression
- `net` module for TCP (Node.js)
- Use streams for efficient processing

## 11. Example Protocol Flow

```
Client                                 Server
  |                                      |
  | --- TCP Connection Establishment --> |
  |                                      |
  | --- TLS Handshake (optional) ------> |
  |                                      |
  | --- Window Size (1024) ------------> |
  |                                      |
  | --- JSON Data (seq=1, count=10) ---> |
  |                                      |
  | <-- Acknowledgment (seq=1) --------- |
  |                                      |
  | --- Compressed Data (seq=2, c=20) -> |
  |                                      |
  | <-- Acknowledgment (seq=2) --------- |
  |                                      |
  | --- Connection Close --------------> |
```

## 12. Binary Examples

### 12.1. Window Size Frame (1024)

```
0x32 0x31                         # Version=2, Type='1'
0x00 0x00 0x00 0x04               # Payload size=4 bytes
0x00 0x00 0x04 0x00               # Window size=1024
```

### 12.2. JSON Data Frame (1 event)

```
0x32 0x32                         # Version=2, Type='2'
0x00 0x00 0x00 0x62               # Payload size=98 bytes
0x00 0x00 0x00 0x01               # Sequence number=1
0x00 0x00 0x00 0x01               # Event count=1
0x00 0x00 0x00 0x56               # JSON length=86 bytes
{ "message": "test log entry", 
  "@timestamp": "2023-05-20T12:00:00.000Z",
  "source": "test.log", 
  "offset": 1234 }
```

### 12.3. Acknowledgment Frame

```
0x32 0x41                         # Version=2, Type='A'
0x00 0x00 0x00 0x04               # Payload size=4 bytes
0x00 0x00 0x00 0x01               # Sequence number=1
```

## 13. Security Considerations

### 13.1. Transport Security

- Use TLS 1.2+ with strong cipher suites
- Validate certificates properly
- Consider mutual TLS for client authentication

### 13.2. Resource Protection

- Implement rate limiting
- Set appropriate buffer limits
- Monitor resource usage

## 14. References

1. Elastic go-lumber implementation: https://github.com/elastic/go-lumber
2. Lumberjack output in Filebeat: https://www.elastic.co/guide/en/beats/filebeat/current/logstash-output.html

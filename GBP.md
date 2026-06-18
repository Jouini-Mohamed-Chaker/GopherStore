# The GopherStore Binary Protocol (GBP)
A Type-Length-Value (TLV) binary protocol for predictable and extremely lean communication between distributed Data Stores.

## 1. The Request Header
Every command sent starts with a strict 7-byte header

Offset,         Field,          Size,           Type,       Description
0,              Opcode,         1 byte,         uint8,      Identifies the command (SET, GET, etc.)
1,              Key Length,     2 bytes,        uint16,     Length of the key in bytes (Max 64KB)
3,              Value Length,   4 bytes,        uint32,     Length of the value in bytes (Max 4GB)

## 2. Opcodes Map
Opcode,     Command,    Expected Payload
0x01,       SET,        key bytes + value bytes
0x02,       GET,        key bytes
0x03,       DEL,        key bytes
0x04,       PING,       none (header only)

### Example: How it Looks in Memory

#### Example A: SET user123 active
Opcode: 0x01
Key Length: 7 bytes (user123) → 0x00 0x07
Value Length: 6 bytes (active) → 0x00 0x00 0x00 0x06
Payload: user123active
Total Bytes Sent: 20 bytes (7 byte header + 13 byte payload).

#### Example B: GET user123
Opcode: 0x02 (GET)
Key Length: 7 bytes (user123) → 0x00 0x07
Value Length: 0 bytes → 0x00 0x00 0x00 0x00
Payload: user123

## 3. Server Response Protocol
Offset,     Field,          Size,       Type,       Description
0,          Status,         1 byte,     uint8,      0x00 (OK), 0x01 (Error), 0x02 (Not Leader)
1,          Value Length,   4 bytes,    uint32      Length of the returned Data or error message

If the client sends a GET, the server replies with a 0x00 status, the length of the value, and the raw bytes of the value.
If the client writes to a follower, the follower returns 0x02, the length of the Leader's IP string, and the IP string itself.
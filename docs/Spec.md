# GopherStore Architecture Specification

## 1. Architectural Overview

The system is a decentralized, cluster-aware key-value store consisting of multiple independent nodes. It utilizes a **Single-Leader Replication** model (Primary-Backup) with a custom TCP-based protocol for speed and control.

```
                  +------------------+
                  |    Client CLI    |
                  +------------------+
                           |
            TCP (Custom Binary Protocol)
                           |
                           v
                  +------------------+
                  |   Leader Node    | (Port 8000)
                  +------------------+
                    /              \
         Replication Stream     Replication Stream
                  /                  \
                 v                    v
      +------------------+    +------------------+
      |  Follower Node 1 |    |  Follower Node 2 | (Port 8001/8002)
      +------------------+    +------------------+

```

---

## 2. Component Specification

### Storage Engine (In-Memory Core)

* **Data Structure:** A native Go `map[string][]byte`. Keys are strings; values are raw byte slices to ensure any arbitrary data type (JSON, strings, binary) can be stored.
* **Concurrency Control:** A `sync.RWMutex` protecting the map.
* *Reads* acquire a shared read lock (`RLock`), allowing high concurrency for lookup operations.
* *Writes* acquire an exclusive write lock (`Lock`), blocking all other operations during modifications.



### Networking & Transport Layer

* **Protocol:** Raw TCP sockets using Go’s `net` package. Avoid HTTP/REST to minimize overhead and maximize throughput.
* **Framing:** A custom binary protocol or custom line-delimited text protocol (similar to Redis RESP).
* *Example Line Protocol:* `*<number of arguments>\r\n$<length of arg1>\r\n<arg1>\r\n...`
* This forces you to manage buffer reading, handle partial TCP packets, and avoid message truncation.



### Node Roles & State Machine

Each node tracks its internal cluster state:

* **Leader:** Accepts reads and writes. Coordinates replication to followers. Sends periodic heartbeats.
* **Follower:** Read-only node. Replicates state changes received from the leader. Monitors leader health via heartbeats.

---

## 3. Protocol & Supported Operations

Your store must support four fundamental network commands:

1. **`SET <key> <value>`**
* *Behavior:* Inserts or updates a key. If executed on a follower, the follower returns a redirection error (`ERR_NOT_LEADER <leader_ip>`).


2. **`GET <key>`**
* *Behavior:* Retrieves the value. Can be served by the Leader or a Follower (depending on your consistency preference).


3. **`DELETE <key>`**
* *Behavior:* Removes a key from the map.


4. **`PING`**
* *Behavior:* Returns `PONG`. Used for health checks and connection validation.



---

## 4. Technical Decisions & Challenges

### Replication Mechanics (Synchronous vs. Asynchronous)

You must choose how data reaches followers. For this project, implement **Asynchronous Replication with an Oplog (Operation Log)**:

* The leader maintains a sequential log of mutations (e.g., `1: SET foo bar`, `2: DEL baz`).
* Each follower maintains a connection to the leader's replication stream and tracks its own `CommitIndex`.
* *Trade-off:* High performance, but there is a risk of a "replication lag" window where followers serve stale data.

### Cluster Membership & Discovery

To keep things simple but challenging, use a static configuration file (e.g., JSON or YAML) provided at node startup:

* The config lists the addresses of all nodes in the cluster and explicitly designates who the initial Leader is.
* Nodes communicate peer-to-peer using background goroutines to track whether their cluster mates are alive.

### Graceful Shutdown & Context Management

* The system must intercept OS signals (`SIGINT`, `SIGTERM`).
* Upon intercepting a shutdown signal, it must use Go’s `context.Context` to propagate a cancellation signal, close active client connections, flush any replication logs to disk (if implementing persistence), and close network listeners safely without panicking.

---

## 5. Step-by-Step Execution Strategy

To avoid feeling overwhelmed, build the system incrementally in these specific phases:

* **Phase 1: Local Engine.** Build just the in-memory map protected by a mutex, wrapped in a clean interface. Write unit tests ensuring thread safety.
* **Phase 2: Networked Single Node.** Wrap your storage engine in a TCP server loop. Write a client CLI parser that reads from stdin, formats the payload, sends it over TCP, and prints the server response.
* **Phase 3: Replication Protocol.** Modify your write paths (`SET`, `DELETE`) on the server so that when a write occurs, the server loops over a slice of follower connections and broadcasts the exact same raw command bytes to them before updating its local map.
* **Phase 4: Fault Tolerance.** Implement a background worker ticker in followers. If they don't receive a network message from the leader within 5 seconds, log an error indicating leader timeout.

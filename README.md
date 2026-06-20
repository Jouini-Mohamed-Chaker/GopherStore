# GopherStore

A distributed, in-memory key-value store written in Go, built with a custom binary wire protocol instead of HTTP/REST.

This is a personal learning project, built for fun and to understand how systems like Redis and etcd work under the hood — replication, custom binary protocols, and the practical edge cases of raw TCP socket programming in Go.

> **Status: Work in progress.** Not production-ready. Don't put anything in here you'd be sad to lose.

All code in this repository is written by hand. No AI tools are used to write, generate, or autocomplete the implementation — the only place AI assistance comes in is writing and formatting documentation like this README.

---

## What is this?

GopherStore explores the core ideas behind distributed key-value stores: an in-memory data structure protected by proper concurrency control, a single-leader replication model, and a lean custom protocol designed specifically for this project rather than reused off the shelf.

It's being built incrementally and in public, so the commit history doubles as a build log.

### Highlights

- Raw TCP, no HTTP overhead — a custom binary protocol instead of JSON/REST
- In-memory store backed by a `map[string][]byte`, safe for concurrent access
- Single-leader replication — a primary node accepts writes and streams them to read-only followers
- Minimal dependencies — mostly Go's standard library

---

## Architecture

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

- **Leader** — accepts reads and writes, coordinates replication, sends heartbeats.
- **Follower** — read-only, replicates from the leader's operation stream, monitors leader health.
- **Replication** — asynchronous, via an append-only operation log. Followers track their own commit index against the leader's. This trades strict consistency for throughput, meaning there's a small window where a follower can serve stale data.
- **Cluster membership** — static, defined in a config file at startup.

---

## The GopherStore Binary Protocol (GBP)

Instead of a text protocol, GopherStore speaks a small Type-Length-Value (TLV) binary protocol. Every request starts with a fixed-size header, so the server always knows exactly how many bytes to read next — no delimiters, no parsing ambiguity.

Requests carry an opcode identifying the command, followed by the lengths of the key and value, followed by the raw key/value bytes themselves. Responses carry a status byte (success, error, or "not leader") along with the length and bytes of any returned data. The full wire format is documented separately as the protocol stabilizes.

Supported operations: `SET`, `GET`, `DEL`, and `PING`.

---

## Getting started

```bash
git clone git@github.com:Jouini-Mohamed-Chaker/GopherStore.git
cd GopherStore
go build ./...
```

Usage instructions will be fleshed out as the CLI client and cluster config format stabilize.

---

## Why not just use Redis / etcd / something that already exists?

Because the point isn't the destination — it's working through distributed systems, binary protocols, and concurrent Go by hand until they make sense from the inside. If you're looking for something to run in production, this isn't it.

---

## Contributing

This is primarily a personal learning project, but issues, suggestions, and pull requests are welcome if something looks interesting to you. No pressure, no SLAs.

## License

Licensed under the [Apache License 2.0](LICENSE).
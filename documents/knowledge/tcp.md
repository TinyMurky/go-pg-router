
# Foundation: What Is TCP?

  Think of TCP like a phone call, not like sending a letter.

  - A letter (UDP) — you write bytes, send them, and hope they arrive.
  No confirmation, no order guarantee.
  - A phone call (TCP) — you establish a connection first (dial → ring →
   answer), then both sides can talk back and forth in order, and when
  someone hangs up, the other side knows.

  PostgreSQL uses TCP. When your app "connects to Postgres", it is
  literally making a phone call to a port on a server. The conversation
  that follows has a defined script — that script is the PostgreSQL wire
   protocol.

  In Go, a TCP connection is represented by one interface:

  // From the standard library — this is the entire abstraction
  type Conn interface {
      Read(b []byte) (n int, err error)
      Write(b []byte) (n int, err error)
      Close() error
      // ... deadline methods
  }

  That's it. A connection is something you can read bytes from and write
   bytes to. Everything in go-pg-router — parsing, routing, proxying —
  is built on top of this one interface.


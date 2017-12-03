/*
Package rateio provides an io interfaces for rate-limiting.

This can be used to apply rate limiting to any type that implements an io-style interface.

For example, we can use it to restrict the reading rate of a net.Conn:

	type limitedConn struct {
		net.Conn
		io.Reader // Our rate-limited io.Reader for net.Conn
	}

	func (r *limitedConn) Read(p []byte) (n int, err error) {
		return r.Reader.Read(p)
	}

	// ReadLimitConn returns a net.Conn whose io.Reader interface is rate-limited by limiter.
	func ReadLimitConn(conn net.Conn, limiter rateio.Limiter) net.Conn {
		return &limitedConn{
			Conn:   conn,
			Reader: rateio.NewReader(conn, limiter),
		}
	}

Then we can use ReadLimitConn to wrap our existing net.Conn and continue using
the wrapped version in its place.

*/
package rateio

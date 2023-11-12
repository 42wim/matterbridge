/*
Package tcp is used to perform TCP handshake without ACK, which useful for TCP health checking.
HAProxy does this exactly the same, which is:

	1. SYN
	2. SYN-ACK
	3. RST

Why do I have to do this

In most cases when you establish a TCP connection(e.g. via net.Dial), these are the first three packets between the client and server(TCP three-way handshake):

	1. Client -> Server: SYN
	2. Server -> Client: SYN-ACK
	3. Client -> Server: ACK

This package tries to avoid the last ACK when doing handshakes.

By sending the last ACK, the connection is considered established.

However, as for TCP health checking the server could be considered alive right after it sends back SYN-ACK, that renders the last ACK unnecessary or even harmful in some cases.

Benefits

By avoiding the last ACK

	1. Less packets better efficiency
	2. The health checking is less obvious

The second one is essential because it bothers the server less.

This means the application level server will not notice the health checking traffic at all, thus the act of health checking will not be considered as some misbehavior of client.
*/
package tcp

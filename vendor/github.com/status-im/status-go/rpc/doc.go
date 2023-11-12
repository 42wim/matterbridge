/*
Package rpc - JSON-RPC client with custom routing.

Package rpc implements status-go JSON-RPC client and handles
requests to different endpoints: upstream or local node.

Every JSON-RPC request coming from either JS code or any other part
of status-go should use this package to be handled and routed properly.

Routing rules are following:

- if Upstream is disabled, everything is routed to local ethereum-go node
- otherwise, some requests (from the list, see below) are routed to upstream, others - locally.

List of methods to be routed is currently available here: https://docs.google.com/spreadsheets/d/1N1nuzVN5tXoDmzkBLeC9_mwIlVH8DGF7YD2XwxA8BAE/edit#gid=0

Note, upon creation of a new client, it ok to be offline - client will keep trying to reconnect in background.
*/
package rpc

//go:generate autoreadme -f

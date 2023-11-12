### Structure

A Status node is a container of services.
These services are passed to geth and registered with geth as APIs and Protocols.

Status node manages all the services and the geth node.

Status node is managed by `api/geth_backend.go`

So:

`GethBackend` manages `StatusNode`, `StatusNode` manages `GethNode`

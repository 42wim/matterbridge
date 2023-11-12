package params

// Version is defined in VERSION file.
// We set it in loadNodeConfig() in api/backend.go.
var Version string

// GitCommit is a commit hash.
var GitCommit string

// IpfsGatewayURL is the Gateway URL to use for IPFS
var IpfsGatewayURL string

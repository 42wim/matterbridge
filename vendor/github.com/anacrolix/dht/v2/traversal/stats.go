package traversal

type Stats struct {
	// Count of (probably) distinct addresses we've sent traversal queries to. Accessed with atomic.
	NumAddrsTried uint32
	// Number of responses we received to queries related to this traversal. Accessed with atomic.
	NumResponses uint32
}

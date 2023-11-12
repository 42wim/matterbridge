package ens

// maxRetries is the maximum number of attempts we do before giving up
const maxRetries uint64 = 11

// ENSBackoffTimeSec is the step of the exponential backoff
// we retry roughly for 17 hours after receiving the message 2^11 * 30
const ENSBackoffTimeSec uint64 = 30

// The pubsub package provides facilities for the Publish/Subscribe pattern of message
// propagation, also known as overlay multicast.
// The implementation provides topic-based pubsub, with pluggable routing algorithms.
//
// The main interface to the library is the PubSub object.
// You can construct this object with the following constructors:
//
// - NewFloodSub creates an instance that uses the floodsub routing algorithm.
//
// - NewGossipSub creates an instance that uses the gossipsub routing algorithm.
//
// - NewRandomSub creates an instance that uses the randomsub routing algorithm.
//
// In addition, there is a generic constructor that creates a pubsub instance with
// a custom PubSubRouter interface. This procedure is currently reserved for internal
// use within the package.
//
// Once you have constructed a PubSub instance, you need to establish some connections
// to your peers; the implementation relies on ambient peer discovery, leaving bootstrap
// and active peer discovery up to the client.
//
// To publish a message to some topic, use Publish; you don't need to be subscribed
// to the topic in order to publish.
//
// To subscribe to a topic, use Subscribe; this will give you a subscription interface
// from which new messages can be pumped.
package pubsub

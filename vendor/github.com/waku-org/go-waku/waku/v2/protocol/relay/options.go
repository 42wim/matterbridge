package relay

import pubsub "github.com/libp2p/go-libp2p-pubsub"

type publishParameters struct {
	pubsubTopic string
}

// PublishOption is the type of options accepted when publishing WakuMessages
type PublishOption func(*publishParameters)

// WithPubSubTopic is used to specify the pubsub topic on which a WakuMessage will be broadcasted
func WithPubSubTopic(pubsubTopic string) PublishOption {
	return func(params *publishParameters) {
		params.pubsubTopic = pubsubTopic
	}
}

// WithDefaultPubsubTopic is used to indicate that the message should be broadcasted in the default pubsub topic
func WithDefaultPubsubTopic() PublishOption {
	return func(params *publishParameters) {
		params.pubsubTopic = DefaultWakuTopic
	}
}

type relayParameters struct {
	pubsubOpts      []pubsub.Option
	maxMsgSizeBytes int
}

type RelayOption func(*relayParameters)

func WithPubSubOptions(opts []pubsub.Option) RelayOption {
	return func(params *relayParameters) {
		params.pubsubOpts = append(params.pubsubOpts, opts...)
	}
}

func WithMaxMsgSize(maxMsgSizeBytes int) RelayOption {
	return func(params *relayParameters) {
		if maxMsgSizeBytes == 0 {
			maxMsgSizeBytes = defaultMaxMsgSizeBytes
		}
		params.maxMsgSizeBytes = maxMsgSizeBytes
	}
}

func defaultOptions() []RelayOption {
	return []RelayOption{
		WithMaxMsgSize(defaultMaxMsgSizeBytes),
	}
}

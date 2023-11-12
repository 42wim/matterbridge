package protocol

import "golang.org/x/exp/maps"

type PubsubTopicStr = string
type ContentTopicStr = string

type ContentTopicSet map[string]struct{}

func NewContentTopicSet(contentTopics ...string) ContentTopicSet {
	s := make(ContentTopicSet, len(contentTopics))
	for _, ct := range contentTopics {
		s[ct] = struct{}{}
	}
	return s
}

func (cf ContentTopicSet) ToList() []string {
	return maps.Keys(cf)
}

// ContentFilter is used to specify the filter to be applied for a FilterNode.
// Topic means pubSubTopic - optional in case of using contentTopics that following Auto sharding, mandatory in case of named or static sharding.
// ContentTopics - Specify list of content topics to be filtered under a pubSubTopic (for named and static sharding), or a list of contentTopics (in case ofAuto sharding)
// If pubSub topic is not specified, then content-topics are used to derive the shard and corresponding pubSubTopic using autosharding algorithm
type ContentFilter struct {
	PubsubTopic   string          `json:"pubsubTopic"`
	ContentTopics ContentTopicSet `json:"contentTopics"`
}

func (cf ContentFilter) ContentTopicsList() []string {
	return cf.ContentTopics.ToList()
}

func NewContentFilter(pubsubTopic string, contentTopics ...string) ContentFilter {
	return ContentFilter{pubsubTopic, NewContentTopicSet(contentTopics...)}
}

func (cf ContentFilter) Equals(cf1 ContentFilter) bool {
	if cf.PubsubTopic != cf1.PubsubTopic ||
		len(cf.ContentTopics) != len(cf1.ContentTopics) {
		return false
	}
	for topic := range cf.ContentTopics {
		_, ok := cf1.ContentTopics[topic]
		if !ok {
			return false
		}
	}
	return true
}

// This function converts a contentFilter into a map of pubSubTopics and corresponding contentTopics
func ContentFilterToPubSubTopicMap(contentFilter ContentFilter) (map[PubsubTopicStr][]ContentTopicStr, error) {
	return GeneratePubsubToContentTopicMap(contentFilter.PubsubTopic, contentFilter.ContentTopicsList())
}

package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// StatsGetResponse struct.
type StatsGetResponse []object.StatsPeriod

// StatsGet returns statistics of a community or an application.
//
// https://dev.vk.com/method/stats.get
func (vk *VK) StatsGet(params Params) (response StatsGetResponse, err error) {
	err = vk.RequestUnmarshal("stats.get", &response, params)
	return
}

// StatsGetPostReachResponse struct.
type StatsGetPostReachResponse []object.StatsWallpostStat

// StatsGetPostReach returns stats for a wall post.
//
// https://dev.vk.com/method/stats.getPostReach
func (vk *VK) StatsGetPostReach(params Params) (response StatsGetPostReachResponse, err error) {
	err = vk.RequestUnmarshal("stats.getPostReach", &response, params)
	return
}

// StatsTrackVisitor adds current session's data in the application statistics.
//
// https://dev.vk.com/method/stats.trackVisitor
func (vk *VK) StatsTrackVisitor(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("stats.trackVisitor", &response, params)
	return
}

// +build !ios,!android

package metrics

import "github.com/elastic/gosigar"

func readCPUStats(stats *CPUStats) {
	global := gosigar.Cpu{}
	global.Get()

	stats.GlobalTime = int64(global.User + global.Nice + global.Sys)
	stats.GlobalWait = int64(global.Wait)
	stats.LocalTime = getProcessCPUTime()
}

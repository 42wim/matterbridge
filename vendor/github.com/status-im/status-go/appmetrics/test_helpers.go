package appmetrics

import "encoding/json"

func GenerateMetrics(num int) []AppMetric {
	var appMetrics []AppMetric
	for i := 0; i < num; i++ {
		am := AppMetric{
			Event:      NavigateTo,
			Value:      json.RawMessage(`{"view_id": "some-view-id", "params": {"screen": "login"}}`),
			OS:         "android",
			AppVersion: "1.11",
		}
		if i < num/2 {
			am.Processed = true
		}
		appMetrics = append(appMetrics, am)
	}

	return appMetrics
}

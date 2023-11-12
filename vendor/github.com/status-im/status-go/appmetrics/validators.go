package appmetrics

var NavigateToCofxSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"view_id": map[string]interface{}{
			"type":      "string",
			"maxLength": 32,
		},
		"params": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"screen": map[string]interface{}{
					"type":      "string",
					"maxLength": 32,
				},
			},
			"additionalProperties": false,
			"required":             []string{"screen"},
		},
	},
	"additionalProperties": false,
	"required":             []string{"view_id", "params"},
}

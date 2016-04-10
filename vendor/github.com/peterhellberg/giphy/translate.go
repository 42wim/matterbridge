package giphy

import (
	"encoding/json"
	"strings"
)

// Translate returns a translate response from the Giphy API
func (c *Client) Translate(args []string) (Translate, error) {
	argsStr := strings.Join(args, " ")

	req, err := c.NewRequest("/gifs/translate?s=" + argsStr)
	if err != nil {
		return Translate{}, err
	}

	var translate Translate
	if _, err = c.Do(req, &translate); err != nil {
		return Translate{}, err
	}

	if len(translate.RawData) == 0 {
		return Translate{}, ErrNoRawData
	}

	// Check if the first character in Data is a [
	if translate.RawData[0] == '[' {
		return Translate{}, ErrNoImageFound
	}

	err = json.Unmarshal(translate.RawData, &translate.Data)
	if err != nil {
		return Translate{}, err
	}

	return translate, nil
}

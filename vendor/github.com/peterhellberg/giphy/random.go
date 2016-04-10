package giphy

import (
	"encoding/json"
	"strings"
)

// Random returns a random response from the Giphy API
func (c *Client) Random(args []string) (Random, error) {
	argsStr := strings.Join(args, " ")

	req, err := c.NewRequest("/gifs/random?tag=" + argsStr)
	if err != nil {
		return Random{}, err
	}

	var random Random
	if _, err = c.Do(req, &random); err != nil {
		return Random{}, err
	}

	// Check if the first character in Data is a [
	if random.RawData == nil || random.RawData[0] == '[' {
		return Random{}, ErrNoImageFound
	}

	var d RandomData

	err = json.Unmarshal(random.RawData, &d)
	if err != nil {
		return Random{}, err
	}

	random.Data = d

	return random, nil
}

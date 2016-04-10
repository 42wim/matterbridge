package giphy

import (
	"fmt"
	"strings"
)

// Search returns a search response from the Giphy API
func (c *Client) Search(args []string) (Search, error) {
	argsStr := strings.Join(args, " ")

	path := fmt.Sprintf("/gifs/search?limit=%v&q=%s", c.Limit, argsStr)
	req, err := c.NewRequest(path)
	if err != nil {
		return Search{}, err
	}

	var search Search
	if _, err = c.Do(req, &search); err != nil {
		return Search{}, err
	}

	return search, nil
}

package giphy

import "fmt"

// Trending returns a trending response from the Giphy API
func (c *Client) Trending(args ...[]string) (Trending, error) {
	path := fmt.Sprintf("/gifs/trending?limit=%v", c.Limit)
	req, err := c.NewRequest(path)
	if err != nil {
		return Trending{}, err
	}

	var res Trending
	if _, err = c.Do(req, &res); err != nil {
		return res, err
	}

	if len(res.Data) == 0 {
		return res, ErrNoTrendingImagesFound
	}

	return res, nil
}

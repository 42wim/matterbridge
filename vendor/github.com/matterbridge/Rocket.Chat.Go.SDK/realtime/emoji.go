package realtime

func (c *Client) getCustomEmoji() error {
	_, err := c.ddp.Call("listEmojiCustom")
	if err != nil {
		return err
	}

	return nil
}

package realtime

import "fmt"

func (c *Client) StartTyping(roomId string, username string) error {
	_, err := c.ddp.Call("stream-notify-room", fmt.Sprintf("%s/typing", roomId), username, true)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) StopTyping(roomId string, username string) error {
	_, err := c.ddp.Call("stream-notify-room", fmt.Sprintf("%s/typing", roomId), username, false)
	if err != nil {
		return err
	}

	return nil
}

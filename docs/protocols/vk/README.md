# VK

- Status: ???
- Maintainers: ???
- Features: ???

## Configuration

**Basic configuration example:**

```toml
[vk.myvk]
# Access token
Token="Yourtokenhere"
```

## FAQ

### How to create an Access Token for my matterbridge bot?

1. Create new Community
2. Manage > API Usage > Create Token
3. Select scope as in the screenshot

![scope](https://user-images.githubusercontent.com/37155836/105666803-e2cfbf00-5efb-11eb-8d8a-5c674309336f.png)

4. Enable longpoll at version 5.126 and enable Message received event

![](https://user-images.githubusercontent.com/37155836/105666939-36420d00-5efc-11eb-9fa1-418aa9cb4a3f.png)
![](https://user-images.githubusercontent.com/37155836/105667066-7903e500-5efc-11eb-9d31-d5cca5a7aac3.png)

## How to get channel ID for vk for my matterbridge bot?

Start matterbridge in debug mode, then send a message in the chat where the bot is. Look for the chat's `PeerID` (usually a number that starts from `2000000000`).

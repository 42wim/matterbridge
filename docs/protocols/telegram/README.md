# Telegram

- Status: Working
- Maintainers: @selfhoster1312
- Features: ???

## Configuration

> [!TIP]
> For detailed information about telegram settings, see [settings.md](settings.md)

**Basic configuration example:**

```toml
[telegram.mytelegram]
Token="Yourtokenhere"
RemoteNickFormat="({PROTOCOL}) {NICK} "
MessageFormat="HTMLNick"
```

## FAQ

### How to get a token for my bot?

See [account.md](account.md)

### How to retrieve my Telegram chat ID?

> [!WARNING]
> So-called topics in "forums" have a different address format! See
> dedicated FAQ answer about this.

To retrive your ChatId, you can:

- post `/chatId` in the Telegram chat, and the bot will reply the ID
- from a desktop computer, navigate to your channel, and copy the numbers at the end of the URL
- look for `Channel:"-XXXXXXX"` in `config.Message` in matterbridge debug log (where `-XXXXXXX` is your ChatId)

### Does matterbridge support Telegram channels?

"Channels" in telegram are read-only chatrooms where only admin can post. Adding a bot to a channel requires giving it administrator rights. If you feel comfortable doing that, please report if it works.

### Why can't I invite my bot into my channel?

As of September 2025, the Telegram web interface is bugged and will silently error when trying to invite a bot into a (read-only) channel as an ordinary member.

Maybe adding it as an administrator directly works. Adding it from the mobile application works (it will prompt to make the bot administrator instead).

### Limitations

- The Telegram API does not report any changes **when messages are deleted** so Matterbridge is unable to remove any bridged messages after they've been sent (This will render many common spam solutions useless). Use regexp with the `IgnoreMessages=` field to remove any common spam messages. # Telegram 

### Matterbridge is not relaying messages from Telegram

See 
* https://core.telegram.org/bots#privacy-mode
* https://github.com/yagop/node-telegram-bot-api/issues/174#issuecomment-244632667

If your bot is not getting messages:

Disable privacy mode with @Botfather.

* Go to [BotFather](https://t.me/botfather) send /setprivacy.
* Select the username of the bot.
* Select Disable.
* Kick bot from chat if it's already in it.
* Invite bot to chat.

The order is important.

### Matterbridge is not relay messages from other bots

> Bots talking to each other could potentially get stuck in unwelcome loops. To avoid this, we decided that bots will not be able to see messages from other bots regardless of mode.

https://core.telegram.org/bots/faq#why-doesn-39t-my-bot-see-messages-from-other-bots

### Matterbridge is not relaying images/stickers/files from Telegram

Because images/stickers/files are from non-public url's, you'll need to setup a [mediaserver](Mediaserver-setup-(advanced))

### Matterbridge is not relaying messages to Telegram

Did you enable `MessageFormat="HTML"` in your config?  
You could be sending invalid HTML. Set it to `MessageFormat=""`

More info in https://github.com/42wim/matterbridge/blob/master/matterbridge.toml.sample#L836-L838

### Matterbridge is not deleting messages from Telegram to other bridges

Telegram doesn't has "deleted messages" metadata, so we don't know which messages are deleted.

### Users are shown as "unknown user"

Telegram channels will always return unknown users. Telegram groups will show usernames if possible

### How to use MediaConvertTgs with lottie?

This requires the external dependency `lottie`, which can be installed like this:
`pip install lottie cairosvg`
https://github.com/42wim/matterbridge/issues/874

Note that if you insist on using an ancient Python version like 3.5, the pip installation is slightly more complicated. Matterbridge expects `lottie_convert.py` to be in your `$PATH`; if that's not already the case, try putting this into your `~/.profile`:
```
PATH=$HOME/.local/bin:$PATH
export PATH
```

If you encounter bugs with this, try to extract the Telegram sticker file and run lottie on it like this:
`lottie_convert.py --input-format lottie file_1234_tgs.webp myoutput.webp`
This might give you additional information about what's going on.

### How to join a topic in a Telegram forum?

What telegram calls a forum is a chatroom which has been split into several sub-rooms, much like a so-called Discord server or a Matrix space.

Each topic then has a `message_thread_id` which is visible from the web interface after the usual ChatId. For example, you may see in your URL `/-XXXXXX/Y` where `-XXXXXX` is your ChatId and `Y` is the topic ID.

Specifically when using topics (not normal groups), you need to add `100` between the `-` and the usual ChatId. For example, `-XXXXXX` becomes `-100XXXXXX`. Then you need to add the topic ID, like so: `-100XXXXXXX/Y`.

> [!WARNING]
> **The first topic created is an exception and should not have `/Y` added to the Channel configuration in your gateway. For this topic only, use `-100XXXXXX`.**

This `100` magic number is to our knowledge not documented in the Telegram Bot API docs, but [has been observed in the wild](https://stackoverflow.com/questions/32423837/telegram-bot-how-to-get-a-group-chat-id) and has cost one matterbridge contributor more than their fair share of mental health points.

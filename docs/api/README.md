# Matterbridge API

If you'd like to develop your bridge in a different language than go, or if you
need to interface with a system which can only perform outgoing connections,
the matterbridge API is here.

> [!TIP]
> For detailed information about API settings, see [settings.md](settings.md)

## OpenAPI specification

Matterbridge API spec can be found on https://app.swaggerhub.com/apis-docs/matterbridge/matterbridge-api/0.1.0-oas3

## Example discord/gitter/api gateway

This example creates a gateway containing 3 bridges: discord,gitter and api.   
The API will listen on localhost port 4242.

### Configure matterbridge.toml

This is based on the example in https://github.com/42wim/matterbridge/wiki/How-to-create-your-config

We first add the API account configuration to this example, this makes the API listen on localhost port 4242 with a buffer of 1000 messages and the nicks we receive will look like "{nick}".

```toml
[api.myapi]
BindAddress="127.0.0.1:4242"
Buffer=1000
RemoteNickFormat="{NICK}"
```

Next we add the API gateway configuration, this make sure that we receive the messages from the other bridges and we can sent them.

```toml
[[gateway.inout]]
account="api.myapi"
channel="api"
```

The full example

```toml
[discord.mydiscord]
Token="MTk4NjIyNDgzNDcdOTI1MjQ4.Cl2FMZ.ZnCjm1XVW7vRze4b7Cq4se7kKWs-abD"
Server="myserver"

[gitter.mygitter]
Token="319fda1761c6875739a489b6772daf2ace4b95d0"

[api.myapi]
BindAddress="127.0.0.1:4242"
Buffer=1000
RemoteNickFormat="{NICK}"

[[gateway]]
name="gateway1"
enable=true

[[gateway.inout]]
account="discord.mydiscord"
channel="general"

[[gateway.inout]]
account="gitter.mygitter"
channel="42wim/mygreatproject"

[[gateway.inout]]
account="api.myapi"
channel="api"
```

### Get the messages from the buffer (GET /api/messages)

If we now type a "test" message in our "general" channel on discord, we get the following result

```bash
$ curl http://localhost:4242/api/messages
```


```json
[
 {
  "text": "test",
  "channel": "general",
  "username": "wim",
  "userid": "227183123686215680",
  "avatar": "https://cdn.discordapp.com/avatars/227183947686215680/bd0e6c7fe63274597a4684884891b79d.jpg",
  "account": "discord.mydiscord",
  "event": "",
  "protocol": "",
  "gateway": "gateway1",
  "parent_id": "",
  "timestamp": "2019-01-09T22:37:18.647108348+01:00",
  "id": "",
  "Extra": null
 }
]
```

After we got this message (and we didn't type anything else), the buffer will be empty and doing a curl will now give an empty result

```bash
$ curl http://localhost:4242/api/messages
```

```json
[]
```

### Stream messages (GET /api/stream)

If we now type a "test" message in our "general" channel on discord, we get the following result

```bash
$ curl http://localhost:4242/api/stream
```

```json
{"text":"","channel":"","username":"","userid":"","avatar":"","account":"","event":"api_connected","protocol":"","gateway":"","parent_id":"","timestamp":"2019-01-09T22:48:33.398737344+01:00","id":"","Extra":null}
{"text":"test","channel":"general","username":"wim","userid":"227183123686215680","avatar":"https://cdn.discordapp.com/avatars/227183947686215680/bd0e6c7fe63274597a4684884891b79d.jpg","account":"discord.mydiscord","event":"","protocol":"","gateway":"gateway1","parent_id":"","timestamp":"2019-01-09T22:48:42.506629373+01:00","id":"","Extra":null}
```

At connect you first get a `api_connected` event, then you'll get a http stream of json messages

### Send message (POST /api/message)

We now post a `test` message from `randomuser` to the gateway `gateway1`

Every bridge on the gateway (gitter and discord) will now receive this message.

```bash
curl -XPOST -H 'Content-Type: application/json'  -d '{"text":"test","username":"randomuser","gateway":"gateway1"}' http://localhost:4242/api/message
```

```json
{"text":"test","channel":"api","username":"randomuser","userid":"","avatar":"","account":"api.local","event":"","protocol":"api","gateway":"gateway1","parent_id":"","timestamp":"2019-01-09T22:53:51.618575236+01:00","id":"","Extra":null}
```

## Security
You can also protect the API with a token by adding a `token` option to your api account configuration.

```toml
[api.myapi]
BindAddress="127.0.0.1:4242"
Buffer=1000
RemoteNickFormat="{NICK}"
Token="verys3cret"
```
If you now want to sent a message with curl you'll have to add a `Authorization: Bearer <token>` header

```bash
curl -H "Authorization: Bearer verys3cret" http://localhost:4242/api/stream
```

## Projects using the API

* [MatterLink](https://github.com/elytra/MatterLink) (Matterbridge link for Minecraft Server chat)
* [pyCord](https://github.com/NikkyAI/pyCord) (crossplatform chatbot)
* [Mattereddit](https://github.com/bonehurtingjuice/mattereddit) (Reddit chat support)

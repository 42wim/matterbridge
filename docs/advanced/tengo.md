Tengo can be used for advanced scripting in matterbridge, you can modify incoming and outgoing messages.

You can also use it to modify the `{TENGO}` parameter in the `RemoteNickFormat`.

```
[tengo]
InMessage="yourscript.tengo"
OutMessage="yourotherscript.tengo"
RemoteNickFormat="remotenickformat.tengo"
```

More information about tengo on: https://github.com/d5/tengo/blob/master/docs/tutorial.md and https://github.com/d5/tengo/blob/master/docs/stdlib.md

FIXME: Document support for [dropping messages](https://github.com/42wim/matterbridge/pull/1272).

## InMessage

InMessage allows you to specify the location of a tengo (https://github.com/d5/tengo/) script.  
This script will receive every incoming message and can be used to modify the Username and the Text of that message. 
 
The script will have the following global variables: \
to modify: `msgUsername` and `msgText` \
to read: `msgChannel` and `msgAccount`

The script is reloaded on every message, so you can modify the script on the fly.

Example script can be found in https://github.com/42wim/matterbridge/tree/master/gateway/bench.tengo
and https://github.com/42wim/matterbridge/tree/master/contrib/example.tengo

The example below will check if the text contains blah and if so, it'll replace the text and the username of that message.
```
text := import("text")
if text.re_match("blah",msgText) {
    msgText="replaced by this"
    msgUsername="fakeuser"
}
```

`InMessage="example.tengo"`

## OutMessage

OutMessage allows you to specify the location of the script that will be invoked on each message being sent to a bridge and can be used to modify the Username and the Text of that message.

The script will have the following global variables: \
read-only: \
`inAccount`, `inProtocol`, `inChannel`, `inGateway`, `inEvent` \
`outAccount`, `outProtocol`, `outChannel`, `outGateway`, `outEvent` \

read-write: \
`msgText`, `msgUsername`

The script is reloaded on every message, so you can modify the script on the fly.

The default script in <https://github.com/42wim/matterbridge/tree/master/internal/tengo/outmessage.tengo>
is compiled in and will be executed if no script is specified.

`OutMessage="example.tengo"`

## RemoteNickFormat

RemoteNickFormat allows you to specify the location of a tengo (https://github.com/d5/tengo/) script.

The script will have the following global variables: \
to modify: `result` \
to read: `channel`, `bridge`, `gateway`, `protocol`, `nick`

The result will be set in `{TENGO}` in the RemoteNickFormat key of every bridge where `{TENGO}` is specified

The script is reloaded on every message, so you can modify the script on the fly.

Example script can be found in <https://github.com/42wim/matterbridge/tree/master/contrib/remotenickformat.tengo>

`RemoteNickFormat="remotenickformat.tengo"`

## More information

You can find more tengo examples/info here:
- <https://github.com/42wim/matterbridge/issues?q=is%3Aissue+tengo+label%3Atengo+is%3Aclosed>
- <https://github.com/42wim/matterbridge/tree/master/internal/tengo>
- <https://github.com/42wim/matterbridge/tree/master/contrib>
- <https://github.com/42wim/matterbridge/wiki/settings#tengo>

The tengo documentation can be found here: <https://github.com/d5/tengo/tree/master/docs>
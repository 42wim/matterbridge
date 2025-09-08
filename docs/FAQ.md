# Matterbridge FAQ

## Can matterbridge relay group messages between public chats on different platforms?

Yes, that's precisely what matterbridge is for.

## Does matterbridge support private chatrooms too?

Yes, but actual support depends on the platform.

## Can matterbridge be used to send private messages between users?

No, that is currently unsupported at the moment. See tracking issue [#191](https://github.com/42wim/matterbridge/issues/191).

## Message being skipped

(From https://github.com/42wim/matterbridge/discussions/1566)

Matterbridge does not relay messages from the user account that it is logged in as. Make sure it has its own account.


   * [Support bridging between any protocols](#support-bridging-between-any-protocols)
   * [Support multiple gateways(bridges) for your protocols](#support-multiple-gatewaysbridges-for-your-protocols)
   * [Message edits and deletes](#message-edits-and-deletes)
   * [Attachment / files handling](#attachment--files-handling)
   * [Username and avatar spoofing](#username-and-avatar-spoofing)
   * [Private groups](#private-groups)
   * [API](#api)

# Support bridging between any protocols
Supported protocols are Discord, Gitter, IRC, Mattermost, Matrix, RocketChat, Slack, Steam, Telegram, XMPP (and our own rest API)   
You can bridge between any (number) protocols. Even the same protocol!

For example you can bridge Discord, Gitter, Matrix and Slack.    
A message typed in Slack will be shown in Discord, Gitter and Matrix (and the other way around)

Follow the steps in [[How-to-create-your-config]]

# Support multiple gateways(bridges) for your protocols
A gateway here is a collection of protocols that you want to bridge.

For example you want to bridge the channel "community" on discord, mattermost and slack.    
But you also want to bridge the channel "development" on discord, mattermost and slack.   

To create the setup above you'll have to create 2 gateways.

See Gateway Config Advanced TODO
# Message edits and deletes
* Support incoming and outgoing edits and deletes: Discord, Mattermost, Slack, Matrix and Telegram.
* Support only incoming edits: Gitter. (gitter API doesn't support outgoing edits)
* Support only incoming edits/deletes: RocketChat
* Support only edits: XMPP
* Support no deletes or edits: IRC, Steam(NR)

NR = Not researched

# Attachment / files handling
The logic is: Public links > Native file upload > External mediaserver > Private links

This means that bridges:
- will receive the "public link" from protocols that have links to files without authentication (Gitter, IRC, Discord)
- that support native file uploads (Discord,Matrix,Mattermost,Slack,Telegram,Rocketchat) will receive a uploaded copy of the file from protocols that have links to files with authentication (Matrix, Mattermost, Slack, Telegram)
- that do NOT support native file uploads (Gitter, IRC, Steam, XMPP) will receive the "public link" from your "external mediaserver" from protocols that have links to files with authentication (Matrix, Mattermost, Slack, Telegram). (If you have configured an "external mediaserver")
- that do NOT support native file uploads (Gitter, IRC, Steam, XMPP) will receive the "private link" from protocols that have links to files with authentication (Matrix, Mattermost, Slack, Telegram). (If you have not configured an "external mediaserver")


For example if you bridge between slack and discord:   
  
When you upload a file to slack. The link you get from slack needs authentication, so matterbridge downloads this file and reuploads it to the destination bridge Discord (which support native file uploads). If you upload a file to discord, matterbridge will just sent the link to slack (because the discord links are public and don't need authentication)

# Username and avatar spoofing
Username spoofing (so it looks like the remote users) only works with webhooks for Discord, Mattermost, Slack.   
Matterbridge reads avatars from Discord, Gitter, Mattermost and Slack but can only send them to Discord, Mattermost and Slack when webhooks are enabled. (Gitter doesn't allow avatar changes). 

To have avatars from mattermost visible on the other bridges, you'll have to use the external mediaserver setup. This is because avatar images from mattermost aren't publicly visible when not logged into mattermost. This means we need to download the avatar images, and upload them to the public mediaserver, and use this (now public) avatar URL in the messages sent to another bridge, eg slack.

See webhook in advanced for specific bridge information TODO

# Private groups
Private groups are supported for Mattermost and Slack

# API
See [[Api]]


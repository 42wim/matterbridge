<!-- TOC -->

- [MSteams - mattermost - matterbridge integration](#MSteams---mattermost---matterbridge-integration)
- [Go to Microsoft Azure portal](#go-to-microsoft-azure-portal)
- [Go to App registrations](#go-to-app-registrations)
- [Create a new App registration](#create-a-new-app-registration)
- [Set Permissions](#set-permissions)
    - [Click on View API Permissions (at the bottom)](#click-on-view-api-permissions-at-the-bottom)
    - [Actually set permissions](#actually-set-permissions)
    - [Wait and let an admin consent them](#wait-and-let-an-admin-consent-them)
    - [Consent](#consent)
    - [Accept permissions](#accept-permissions)
    - [Wait again](#wait-again)
    - [Reload](#reload)
- [Set redirect URI](#set-redirect-uri)
- [Set application as public client](#set-application-as-public-client)
- [Get necessary ID's for matterbridge](#get-necessary-ids-for-matterbridge)
    - [ClientID and TenantID](#clientid-and-tenantid)
    - [TeamID](#teamid)
    - [ChannelID](#channelid)
- [Matterbridge configuration](#matterbridge-configuration)
    - [Configure teams in matterbridge](#configure-teams-in-matterbridge)
    - [Configure mattermost in matterbridge](#configure-mattermost-in-matterbridge)
    - [Configure bridging channels](#configure-bridging-channels)
    - [Once again the complete configuration](#once-again-the-complete-configuration)
    - [Starting matterbridge](#starting-matterbridge)

<!-- /TOC -->

# MSteams - mattermost - matterbridge integration

This is a complete walkthrough about how to setup an example mattermost <=> microsoft teams integration using matterbridge.

Please read everything very careful!

# Go to Microsoft Azure portal

- https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/Overview

# Go to App registrations

- https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps

![image](https://user-images.githubusercontent.com/1810977/71424191-ba89e700-268f-11ea-9733-a0fb193c1fb7.png)

# Create a new App registration

Click on `New Registration` (top)

![image](https://user-images.githubusercontent.com/1810977/71424288-7f3be800-2690-11ea-93c6-ee4d811e2bdf.png)

# Set Permissions

## Click on View API Permissions (at the bottom)

![image](https://user-images.githubusercontent.com/1810977/71424300-9bd82000-2690-11ea-9818-6103c09dbbd6.png)

## Actually set permissions

- Choose `graph API` 
- Choose `delegated permissions`
- Add `Group.Read.All`, `Group.ReadWrite.All` and `User.Read`. These permissions are needed for sending/reading chat messages in a channel.
- Add `Files.Read`, `Files.Read.All`, `Sites.Read.All`. These permissions are needed for reading the file attachments in messages.

![image](https://user-images.githubusercontent.com/1810977/71424310-b4483a80-2690-11ea-8c28-051694f2972a.png)

## Wait and let an admin consent them

This can take a while according to the message

![image](https://user-images.githubusercontent.com/1810977/71424323-d04bdc00-2690-11ea-961a-5963f8d02e97.png)

## Consent

You can now click on the Grant admin consent for `yourorganization`

![image](https://user-images.githubusercontent.com/1810977/71424329-df328e80-2690-11ea-8f2c-4e679f5d0460.png)

## Accept permissions

You'll get a popup with the permissions you just added. Agree

![image](https://user-images.githubusercontent.com/1810977/71424341-ef4a6e00-2690-11ea-94ac-1dd7d7737ce0.png)

## Wait again

This will take a few minutes again :)

![image](https://user-images.githubusercontent.com/1810977/71424347-fc675d00-2690-11ea-90cb-a9c95a482d78.png)

## Reload

Afterwards you'll see green checkboxes for the permissions

![image](https://user-images.githubusercontent.com/1810977/71424355-0c7f3c80-2691-11ea-9cd5-a91c7fd0ae3b.png)

# Set redirect URI

This needs to be set otherwise the delegation doesn't work. Click on "Add a redirect URI"

![image](https://user-images.githubusercontent.com/1810977/71424361-1b65ef00-2691-11ea-9b0e-e8bf271451d8.png)

Just fill in something like http://localhost:12345/matterbridge

![image](https://user-images.githubusercontent.com/1810977/71424371-2de02880-2691-11ea-8ce8-fa4535e7468d.png)

# Set application as public client

Scroll down a bit

Set `Treat application as a public client.` to Yes

![image](https://user-images.githubusercontent.com/1810977/71424383-3cc6db00-2691-11ea-94e3-d17fb6faee11.png)

Don't forget to click Save on top of the page


# Get necessary ID's for matterbridge

## ClientID and TenantID

Click on overview, left upper link.

You'll see 2 ID's, these are needed for the matterbridge configuration. 

- Tenant ID
- Client ID

![image](https://user-images.githubusercontent.com/1810977/71424388-4b14f700-2691-11ea-88fa-e8bcaeeb6747.png)

## TeamID

Go to your teams website <https://teams.microsoft.com> should work.

Find your team, click on the 3 dots and select `get link to team`

![image](https://user-images.githubusercontent.com/1810977/71424396-5c5e0380-2691-11ea-853f-182ae192f787.png)

This will get you a popup, click copy.

![image](https://user-images.githubusercontent.com/1810977/71424402-697af280-2691-11ea-9054-30edce6b9e0a.png)

If you paste it you'll get something like

https://teams.microsoft.com/l/team/19%3axxxxxxxxxxxxxxxxxc%40thread.skype/conversations?groupId=**xxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx**&tenantId=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

The groupID is the TeamID you need to configure matterbridge.

## ChannelID

Next you have to decide which channel you want to bridge with for example mattermost.

In our setup we have the team `matterbridge` with 2 channels `General` (a default channel for every team) and `newchannel` one I created.

You'll find the channel ID in the URL in the `threadId=`**19:82abcxxxxxxxxx@thread.skype**

![image](https://user-images.githubusercontent.com/1810977/71424405-7861a500-2691-11ea-9567-c595efe07818.png)

Note this ID **19:82abcxxxxxxxxx@thread.skype**, we will need it when configuring the bridging.


# Matterbridge configuration

Create an empty `matterbridge.toml` file

## Configure teams in matterbridge

You should know have all the three ID's to configure matterbridge:

```toml
[msteams.teams]
TenantID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" 
ClientID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
TeamID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
```

add this to the `matterbridge.toml` file

## Configure mattermost in matterbridge

See als the [wiki](https://github.com/42wim/matterbridge/wiki/Section-Mattermost-(basic)-https)

Configure this for your setup and add this to the `matterbridge.toml` file


```toml
[mattermost.mymattermost]
#The mattermost hostname. (do not prefix it with http or https)
Server="yourmattermostserver.domain:443"

#the team name as can be seen in the mattermost webinterface URL
#in lowercase, without spaces
Team="yourteam"

#login/pass of your bot.
#Use a dedicated user for this and not your own!
Login="yourlogin"
Password="yourpass"

RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
PrefixMessagesWithNick=true
```

## Configure bridging channels

If you want to bridge the `testing` channel in mattermost with the `general` channel in msteams the configuration will look like this:

```toml
[[gateway]]
name="gw"
enable=true

[[gateway.inout]]
account = "mattermost.mymattermost"
channel = "testing"

[[gateway.inout]]
account="msteams.teams"
channel="19:82caxxxxxxxxxxxxxxxxxxxxxxxx@thread.skype"
```

The strange channel **19:82caxxxxxxxxxxxxxxxxxxxxxxxx@thread.skype** can be found in this documentation at the **ChannelID** header above.

## Once again the complete configuration

Your `matterbridge.toml` file should contain:

```toml
[msteams.teams]
TenantID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" 
ClientID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
TeamID="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "

[mattermost.mymattermost]
Server="yourmattermostserver.domain:443"
Team="yourteam"
Login="yourlogin"
Password="yourpass"
RemoteNickFormat="[{PROTOCOL}] <{NICK}> "
PrefixMessagesWithNick=true

[[gateway]]
name="gw"
enable=true

[[gateway.inout]]
account = "mattermost.mymattermost"
channel = "testing"

[[gateway.inout]]
account="msteams.teams"
channel="19:82caxxxxxxxxxxxxxxxxxxxxxxxx@thread.skype"
```


## Starting matterbridge

Now you can start matterbridge by running `matterbridge -conf matterbridge.toml`

The first time you start matterbridge it'll ask you to authenticate the app on behalf of you. You can do this from your own account or use a specific bot account for it. 

Matterbridge can only read/send to the channels the account is in

```bash
[0003]  INFO router:       Starting bridge: msteams.teams
To sign in, use a web browser to open the page https://microsoft.com/devicelogin and enter the code C8EGY6384 to authenticate.
```

Go to the URL as specified and enter the code.

![image](https://user-images.githubusercontent.com/1810977/71424412-8b747500-2691-11ea-92ee-294e82a1fdf1.png)

You'll now get a popup to consent, this is everything that matterbridge has access to. For now it'll only use read all groups and read and write all groups to read and send messages.

![image](https://user-images.githubusercontent.com/1810977/71424430-b5c63280-2691-11ea-8aa9-ababae5b7a6a.png)

Afterwards you should see this window

![image](https://user-images.githubusercontent.com/1810977/71424441-c8406c00-2691-11ea-8ead-ce725875dea9.png)

And matterbridge will continue to start-up

Matterbridge by default will write a sessionfile containing tokens to the directory where matterbridge is running. It'll be a file called `msteams_session.json`. This files contains the necessary credentials so that matterbridge can restart/renew without asking the device login again.

Be sure to keep this file secure!

You can choose another path/filename, by adding `SessionFile="yourfilename"` to the `[msteams.teams]` configuration.
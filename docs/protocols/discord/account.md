(based upon https://github.com/reactiflux/discord-irc/wiki/Creating-a-discord-bot-&-getting-a-token)

## Create Bot

First, you need to go [here](https://discord.com/developers/applications) and click "New Application".

![Application Screen](https://user-images.githubusercontent.com/987487/218297892-e84b1b63-d639-4c26-9b7b-93c809c996a6.png)

Now give your bot a name.

![Create an Application Screen](https://user-images.githubusercontent.com/987487/218297854-2a7aa14b-b8f5-45d2-be39-db6253e4a596.png)

Click "Create". On the next screen, you can optionally set an Avatar Icon for your app and description.

![New App Screen](https://user-images.githubusercontent.com/987487/212194341-b8b7fca2-7b70-422c-b919-09c58a6a0495.png)

Next, click on Bot on the left-hand menu

![Bot menu item](https://user-images.githubusercontent.com/987487/218298058-f61f44d3-aba7-4639-9987-e52bf60b34dc.png)

hen click "Add Bot"

![Add Bot](https://user-images.githubusercontent.com/987487/212194343-5ca1e46e-8e59-4f5a-8f26-08e4894f42e0.png)

Click "Yes Do It"

![Yes Do It.](https://user-images.githubusercontent.com/987487/212194345-56c1b9cb-63cb-4216-b705-2ba1baa6d4a6.png)

On the Bot screen you will generate your **token**. Click "Reset Token" and "Yes, Do It!" a second time.

![Reset Token](https://user-images.githubusercontent.com/987487/212194346-86dcb128-4722-4627-acfc-2d18bb8a8ed4.png)

![Yes, Reset Token](https://user-images.githubusercontent.com/987487/212194347-0c424124-d5c2-4939-8f25-a492a6f6c79b.png)

Then you can click "Copy" to put the token it in your clipboard. Keep this for your matterbridge config file. 

![Token](https://user-images.githubusercontent.com/987487/212194348-d15c249f-d8e7-45ea-b807-ddedaaeb6442.png)

Here, you can also toggle if the bot is public, which will allow others to invite to their servers. This is off by default and will not do anything unless the matterbridge configuration file lists the servers it has been invited to, so you should just leave this off.

![Public Bot](https://user-images.githubusercontent.com/987487/218297521-2290e231-204b-4a68-8044-0879b3c84ce3.png)


### Privileged Gateway Intents

Make sure to also toggle the "Server Members Intent" and "Message Content Intent" options further down under "Privileged Gateway Intents" to allow the bot to see the member list, otherwise you'll get an error message similar to `Error obtaining server members: HTTP 403 Forbidden, {"message": "Missing Access", "code": 50001}`.

Click "Save Changes"

![Server Members Intent](https://user-images.githubusercontent.com/987487/212194349-c2544381-5009-4c90-8d9d-1ce5dcfae855.png)

## Invite bot

Now it's time to invite your bot to your server. Don't worry about your bot being started for this next step. You can get this under OAuth2 → General
![Client ID](https://user-images.githubusercontent.com/987487/212194350-88d037a8-e688-4c43-b2ee-ec41a1990b79.png)

Copy the client ID into this URL and go to it

```https://discordapp.com/oauth2/authorize?&client_id=YOUR_CLIENT_ID_HERE&scope=bot&permissions=536870912```

The weird number (536870912 = 0x20000000) corresponds to the "Manage Webhooks" permission. If you don't want to use `AutoWebhooks=true`, then you can use `0` instead, but you will need to configure the necessary webhooks manually.

It will prompt you to select your server

![Add Bot](https://user-images.githubusercontent.com/987487/212195690-97e5c675-b34f-4dba-895e-232eedb2f7f7.png)

and then Authorize it

![Authorize Bot](https://user-images.githubusercontent.com/987487/212195594-6c709139-38fa-4ea9-a153-95fa55e64363.png)

Solve CAPTCHA if necessary

![2020-10-14-133937_317x178_scrot](https://user-images.githubusercontent.com/987487/96025132-b8616680-0e22-11eb-983b-a9e10426965f.png)

Then you should see

![Authorized](https://user-images.githubusercontent.com/987487/212195597-dc4b9a47-87ff-4a03-90b4-b2db034bf606.png)

You can confirm the bot is added by looking in your #general channel for "A wild [bot name] appeared!" or "Welcome [bot name]. We hope you brought pizza." or some other pithy welcome message, and checking under Your Guild > Settings > Integrations > Bots and Apps.


## Invite to channels

For each channel you want to bridge, you need to make sure the bot user is a member.

The bot should be in all public channels in your server by default, but for private channels you must either make sure to give it an appropriate Role, or to directly add it under Edit Channel > Permissions > Roles/Members

![Edit Channel](https://user-images.githubusercontent.com/987487/218298368-a236ede4-c08a-46df-901c-96a7c33b6b13.png)
![Permissions](https://user-images.githubusercontent.com/987487/218298370-70078ee3-bf80-4082-9f72-c6510fd8d99d.png)

It needs to have View and Send permissions

![View Permission](https://user-images.githubusercontent.com/987487/218298416-c5c31b4a-0b90-4bdb-b632-eddb15e5a158.png)

![Send Permission](https://user-images.githubusercontent.com/987487/218298425-dff62127-b27c-44e6-9285-66ad06f94ea3.png)


If you have trouble with this, open an issue and we can improve these docs.


## Matterbridge Config

Take the bot's token (the _token_, not the client ID) and add it to your matterbridge config like so:

```
[discord.yourbridgeid]
Token="MY_SECRET_TOKEN"
# ....
``` 

Then [[set up your individual bridges|Section-Discord-(basic)]].
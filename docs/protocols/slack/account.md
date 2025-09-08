# A Matterbridge integration for your Slack Workspace

### Contents

[Bot-based Setup](#bot-based-setup)<br/>
<strike>[Legacy Setup](#legacy-setup)</strike><br/>
[Issues](#issues)

---

# Bot-based Setup
Slack's model for bot users and other third-party integrations revolves around Slack Apps. They have been around for a while and are the only and default way of integrating services like Matterbridge going forward.

Slack Docs: [Bot user tokens](https://api.slack.com/docs/token-types#bot)

## Create the **classic** Slack App
> **It's important that you create a "Slack App (Classic)". Don't use the "Create New App" button, and don't opt in to the "granular permissions" feature.**

1. Navigate to Slack's ["Your Apps" page](https://api.slack.com/apps) and log into an account that has administrative permissions over the Slack Workspace (server) that you want to sync with Matterbridge.

   <img alt="Slack's 'Your Apps' page" width=780 src="https://user-images.githubusercontent.com/22248748/119397026-9f1b7500-bca3-11eb-98e6-b3ea925333de.png" />

2. Use [_**THIS LINK**_](https://api.slack.com/apps?new_classic_app=1) to create a "Slack App (Classic)". Choose any name for your app and select the desired workspace, and then submit. :warning: _**USE THE LINK AND DON'T CLICK THE "Create New App" BUTTON**_.

   <img alt="Create a 'Slack App (Classic)'" width=780 src="https://user-images.githubusercontent.com/22248748/119397031-9f1b7500-bca3-11eb-9eaa-cc2418adb9db.png" />

3. Navigate to the "App Home" page via the menu on the left. Click "Add Legacy Bot User", fill in the bot's details, and submit.

   <img alt="Add Legacy Bot User" width=780 src="https://user-images.githubusercontent.com/22248748/119397034-9fb40b80-bca3-11eb-8187-df62925c24f1.png" />

## Grant scopes and install the Slack App

You'll need to give your new Slack App (Classic), and thus the bot, the right permissions on your Slack Workspace.

1. Click "OAuth & Permissions" in the menu on the left, and scroll down to the "Scopes" section. Don't click the prominent "Update Scopes" button. Instead, click "Add an OAuth Scope".

   <img alt="Add an OAuth Scope" width=780 src="https://user-images.githubusercontent.com/22248748/119397023-9dea4800-bca3-11eb-9957-c780e96f0960.png" />

2. Add the following permission scopes to your app:
   * Modify your public channels (`channels:write`)
   * Send messages as \<App name\> (`chat:write:bot`)
   * Send messages as user (`chat:write:user`)
   * Access user’s profile and workspace profile fields (`users.profile:read`)

   <img alt="App with scopes" width=780 src="https://user-images.githubusercontent.com/22248748/119397039-a17dcf00-bca3-11eb-9bf4-25cd24f96aaf.png" />

3. Scroll to the top of the same "OAuth & Permissions" page and click on the "Install to Workspace" (or "Reinstall to Workspace") button:

   <img alt="Install to Workspace" width=780 src="https://user-images.githubusercontent.com/22248748/119397036-a04ca200-bca3-11eb-99a8-545d74effde0.png" />

Confirm that the authorizations you just added are OK:

   <img alt="Confirm install" width=780 src="https://user-images.githubusercontent.com/22248748/119397035-a04ca200-bca3-11eb-9cf2-edac292fa06c.png" />

4. Once the App has been installed, the top of the `OAuth & Permissions` page will show two tokens: a "User OAuth Token" starting with `xoxp-...` and a "Bot User OAuth Token" starting with `xoxb-...`. You will need to use the second in your Matterbridge configuration:

   <img alt="Bot User OAuth Token" width=780 src="https://user-images.githubusercontent.com/22248748/119397029-9f1b7500-bca3-11eb-9faa-0978a0b280e8.png" />

## Invite the bot to channels synced with Matterbridge

The only thing that remains now is to set up the newly created bot on the Slack Workspace itself.

1. On your Slack server you need to add the newly created bot to the relevant channels. Simply use the `/invite @<botname>` command in the chatbox.

   <img width="527" alt="Invite the bot user to a channel" src="https://user-images.githubusercontent.com/22248748/119407127-c9742f00-bcb1-11eb-870d-7e38335f58df.png">

2. Repeat the invite process for each channel that Matterbridge needs to sync. :warning: Also, don't forget to do this in the future when you want to sync more channels.

Now you are all set to go. Just configure and start your Matterbridge instance and see the messages flowing.

   <img width="487" alt="Hello from Zulip!" src="https://user-images.githubusercontent.com/22248748/119406733-320edc00-bcb1-11eb-9e20-28ecd0b98d5d.png">


# Legacy setup

Update: 2020-04-03 - this setup is not working anymore. (See https://github.com/42wim/matterbridge/issues/964#issuecomment-607629250)

🛑 This setup is not recommended and will disappear in future versions of Matterbridge. Please use it only if you are absolutely sure that you can not use the normal setup. Legacy tokens are, as their name indicates, a legacy feature and Slack can in the future **deprecate them at short notice** as they have a weak security model and **give access to a wide-range of privileged actions**. Always store your token securely!

Slack Docs: [Legacy tokens](https://api.slack.com/docs/token-types#legacy)

***

In order to use this setup your channel configuration should use the `slack-legacy` setup instead of `slack`:
```toml
[slack-legacy]
[slack-legacy.myslack]
Token="xoxp-123456789-mytoken"
```

Steps to follow:
1. First create a dedicated user (a new account specifically for the bot) in Slack.

2. Next log in as this user and go to [the legacy tokens page](https://api.slack.com/custom-integrations/legacy-tokens) and click "Create token" for your team and your new user.
   ![Create token screen](https://i.snag.gy/vpsSM4.jpg)
   The token will look something like ```xoxp-123456789-LowerAndUPpercase```.

3. Use the obtained token in your configuration file for the channel setup as in the example above.

# Issues

If you get the message: `ERROR slack: Connection failed "not_allowed_token_type" &errors.errorString{s:"not_allowed_token_type"}`

For more information look at https://github.com/slackapi/node-slack-sdk/issues/921#issuecomment-570662540 and https://github.com/nlopes/slack/issues/654 and our issue https://github.com/42wim/matterbridge/issues/964


### v0.6.0 - August 31, 2019
full differences can be viewed using `git log --oneline --decorate --color v0.5.0..v0.6.0`
thanks to everyone who has contributed since January!


#### Breaking Changes:
- Info struct has had fields removed related to deprecated functionality by slack.
- minor adjustments to some structs.
- some internal default values have changed, usually to be more inline with slack defaults or to correct inability to set a particular value. (Message Parse for example.)

##### Highlights:
- new slacktest package easy mocking for slack client. use, enjoy, please submit PRs for improvements and default behaviours! shamelessly taken from the [slack-test repo](https://github.com/lusis/slack-test) thank you lusis for letting us use it and bring it into the slack repo.
- blocks, blocks, blocks.
- RTM ManagedConnection has undergone a significant cleanup.
in particular handles backoffs gracefully, removed many deadlocks,
and Disconnect is now much more responsive.

### v0.5.0 - January 20, 2019
full differences can be viewed using `git log --oneline --decorate --color v0.4.0..v0.5.0`
- Breaking changes: various old struct fields have been removed or updated to match slack's api.
- deadlock fix in RTM disconnect.

### v0.4.0 - October 06, 2018
full differences can be viewed using `git log --oneline --decorate --color v0.3.0..v0.4.0`
- Breaking Change: renamed ApplyMessageOption, to mark it as unsafe,
this means it may break without warning in the future.
- Breaking: Msg structure files field changed to an array.
- General: implementation for new security headers.
- RTM: deadlock fix between connect/disconnect.
- Events: various new fields added.
- Web: various fixes, new fields exposed, new methods added.
- Interactions: minor additions expect breaking changes in next release for dialogs/button clicks.
- Utils: new methods added.

### v0.3.0 - July 30, 2018
full differences can be viewed using `git log --oneline --decorate --color v0.2.0..v0.3.0`
- slack events initial support added. (still considered experimental and undergoing changes, stability not promised)
- vendored depedencies using dep, ensure using up to date tooling before filing issues.
- RTM has improved its ability to identify dead connections and reconnect automatically (worth calling out in case it has unintended side effects).
- bug fixes (various timestamp handling, error handling, RTM locking, etc).

### v0.2.0 - Feb 10, 2018

Release adds a bunch of functionality and improvements, mainly to give people a recent version to vendor against.

Please check [0.2.0](https://github.com/nlopes/slack/releases/tag/v0.2.0)

### v0.1.0 - May 28, 2017

This is released before adding context support.
As the used context package is the one from Go 1.7 this will be the last
compatible with Go < 1.7.

Please check [0.1.0](https://github.com/nlopes/slack/releases/tag/v0.1.0)

### v0.0.1 - Jul 26, 2015

If you just updated from master and it broke your implementation, please
check [0.0.1](https://github.com/nlopes/slack/releases/tag/v0.0.1)

# v0.6.0
## Breaking changes from 0.5 to 0.6
### commandline
* -plus switch deprecated. Use ```Plus=true``` or ```Plus``` in ```[general]``` section
### IRC section
* ```Enabled``` added (default false)
Add ```Enabled=true``` or ```Enabled``` to the ```[IRC]``` section if you want to enable the IRC bridge
### Mattermost section
* ```Enabled``` added (default false)
Add ```Enabled=true``` or ```Enabled``` to the ```[mattermost]``` section if you want to enable the mattermostbridge
### General section
* Use ```Plus=true``` or ```Plus``` in ```[general]``` section to enable the API version of matterbridge

## New features
* Matterbridge now bridges between any specified protocol (not only mattermost anymore) 
* XMPP support added.  See matterbridge.conf.sample for more information
* RemoteNickFormat {BRIDGE} variable added
You can now add the originating bridge to ```RemoteNickFormat```  
eg ```RemoteNickFormat="[{BRIDGE}] <{NICK}> "```

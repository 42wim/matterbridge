Matterbridge as a Service in Linux
==================================

To install Matterbridge as a service in Linux, place matterbridge and the config file in the `/opt/matterbridge` directory, like so:

```
sudo mkdir -p /opt/matterbridge
sudo cp matterbridge-linux64 /opt/matterbridge
sudo cp matterbridge.toml /opt/matterbridge
```

Then, install the [matterbridge deamon script](matterbridge) service into `/etc/init.d/` like so (description for Debian like installations):

```
sudo cp matterbridge /etc/init.d/
sudo chmod +x /etc/init.d/matterbridge
sudo systemctl daemon-reload
sudo systemctl enable matterbridge
sudo systemctl start matterbridge
```

The Matterbridge will now be running, and will restart after a reboot (it will wait for mattermost to start first).

Output and logging can be found in `/var/log/matterbridge.log`

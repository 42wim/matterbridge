# Mumble

- Status: ???
- Maintainers: ???
- Features: ???

## Configuration

**Basic configuration example:**

```toml
[mumble.mymumble]
Server = "mumble.yourdomain.me:64738"
Nick = "matterbridge"
# Use `""` if the server doesn't have a password
Password = "serverpasswordhere"
# Uncomment for client certificates
#TLSClientCertificate="mumble.crt"
#TLSClientKey="mumble.key"
# Optional custom TLS certificate
TLSCACertificate=mumble-ca.crt
```

## FAQ

### How to generate a client certificate to reserve a nickname?

Self-signed TLS client certificate + private key used to connect to
Mumble.  This is required if you want to register the matterbridge
user on your Mumble server, so its nick becomes reserved.
You can generate a keypair using e.g.

     openssl req -x509 -newkey rsa:2048 -nodes -days 10000 \
            -keyout mumble.key -out mumble.crt

To actually register the matterbridege user, connect to Mumble as an
admin, right click on the user and click "Register".

MailServer
==========

This document is meant to collect various information about our MailServer implementation.

## Syncing between mail servers

It might happen that one mail server is behind other due to various reasons like a machine being down for a few minutes etc.

There is an option to fix such a mail server:
1. SSH to a machine where this broken mail server runs,
2. Add a mail server from which you want to sync:
```
# sudo might be not needed in your setup
$ echo '{"jsonrpc":"2.0","method":"admin_addPeer", "params": ["enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@206.189.243.162:30504"], "id":1}' | \
    sudo socat -d -d - UNIX-CONNECT:/docker/statusd-mail/data/geth.ipc
```
3. Mark it as a trusted peer:
```
# sudo might be not needed in your setup
$ echo '{"jsonrpc":"2.0","method":"shh_markTrustedPeer", "params": ["enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@206.189.243.162:30504"], "id":1}' | \
    sudo socat -d -d - UNIX-CONNECT:/docker/statusd-mail/data/geth.ipc
```
4. Finally, trigger the sync command:
```
# sudo might be not needed in your setup
$ echo '{"jsonrpc":"2.0","method":"shhext_syncMessages","params":[{"mailServerPeer":"enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@206.189.243.162:30504", "to": 1550479953, "from": 1550393583, "limit": 1000}],"id":1}' | \
    sudo socat -d -d - UNIX-CONNECT:/docker/statusd-mail/data/geth.ipc
```

You can add `"followCursor": true` if you want it to automatically download messages until the cursor is empty meaning all data was synced.

### Debugging

To verify that your mail server received any responses, watch logs and seek for logs like this:
```
INFO [02-18|09:08:54.257] received sync response count=217 final=false err= cursor=[]
```

And it should finish with:
```
INFO [02-18|09:08:54.431] received sync response count=0 final=true err= cursor=[]
```

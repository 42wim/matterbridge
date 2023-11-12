Mailservers Service
================

Mailservers service provides read/write API for `Mailserver` object 
which stores details about user's mailservers.

To enable this service, include `mailservers` in APIModules:


```json
{
  "MailserversConfig": {
    "Enabled": true
  },
  "APIModules": "mailservers"
}
```

API
---

Enabling service will expose three additional methods:

#### mailservers_addMailserver

Stores `Mailserver` in the database.

```json
{
    "id": "1",
    "name": "my mailserver",
    "address": "enode://...",
    "password": "some-pass",
    "fleet": "prod"
}
```

#### mailservers_getMailservers

Reads all saved mailservers.

#### mailservers_deleteMailserver

Deletes a mailserver specified by an ID.

## Mailserver requests gap service

Mailserver request gaps service provides read/write API for `MailserverRequestGap` object 
which stores details about the gaps between mailserver requests.

API
---

The service exposes four methods

#### mailserverrequestgaps_addMailserverRequestGaps

Stores `MailserverRequestGap` in the database.
All fields are specified below:

```json
{
  "id": "1",
  "chatId": "chat-id",
  "from": 1,
  "to": 2
}
```

#### mailservers_getMailserverRequestGaps

Reads all saved mailserver request gaps by chatID.

#### mailservers_deleteMailserverRequestGaps

Deletes all MailserverRequestGaps specified by IDs.

#### mailservers_deleteMailserverRequestGapsByChatID

Deletes all MailserverRequestGaps specified by chatID.

#### mailservers_addMailserverTopic

Stores `MailserverTopic` in the database.
```json
{
    "topic": "topic-as-string",
    "chat-ids": ["a", "list", "of", "chatIDs"],
    "last-request": 1
}
```

#### mailservers_getMailserverTopics

Reads all saved mailserver topics.

#### mailservers_deleteMailserverTopic

Deletes a mailserver topic using `topic` as an identifier.

#### mailservers_addChatRequestRange

Stores `ChatRequestRange` in the database.
```json
{
    "chat-id": "chat-id-001",
    "lowest-request-from": 1567693421154,
    "highest-request-to": 1567693576779 
}
```

#### mailservers_getChatRequestRanges

Reads all saved chat request ranges.

#### mailservers_deleteChatRequestRange

Deletes a chat request range by `chat-id`.

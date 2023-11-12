# protocol/encryption package

## Hash ratchet encryption
`encryptor.GenerateHashRatchetKey()` generates a hash ratchet key and stores it in in the DB.
There, 2 new tables are created: `hash_ratchet_encryption` and `hash_ratchet_encryption_cache`.
Each hash ratchet key is uniquely identified by the `(groupId, keyId)` pair, where `keyId` is derived from a clock value.

`protocol.BuildHashRatchetKeyExchangeMessage` builds an 1-on-1 message containing the hash ratchet key, given it's ID.

`protocol.BuildHashRatchetMessage` builds a hash ratchet message with arbitrary payload, given `groupId`. It will use the latest hash ratchet key available. `encryptor.encryptWithHR` encrypts the payload using Hash Ratchet algorithms. Intermediate hashes are stored in `hash_ratchet_encryption_cache` table.

`protocol.HandleMessage` uses `encryptor.decryptWithHR` fn for decryption.

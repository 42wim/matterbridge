# Account Generator

The Account Generator is used to generate, import, derive child keys, and store accounts.
It is instantiated in the `account.Manager` struct and it's accessible from the `lib` and `mobile`
package through functions with the `MultiAccount` prefix:

* MultiAccountGenerate
* MultiAccountGenerateAndDeriveAddresses
* MultiAccountImportMnemonic
* MultiAccountDeriveAddresses
* MultiAccountStoreDerivedAccounts
* MultiAccountImportPrivateKey
* MultiAccountStoreAccount
* MultiAccountLoadAccount
* MultiAccountReset


Using `Generate` and `ImportMnemonic`, a master key is loaded in memory and a random temporarily id is returned.
Bare in mind these accounts are not saved. They are in memory until `StoreAccount` or `StoreDerivedAccounts` are called.
Calling `Reset` or restarting the application will remove everything from memory.
Logging-in and Logging-out will do the same.

Since `Generate` and `ImportMnemonic` create extended keys, we can use those keys to derive new child keys.
`MultiAccountDeriveAddresses(id, paths)` returns a list of addresses/pubKey, one for each path.
This can be used to check balances on those addresses and show them to the user.

Once the user is happy with some specific derivation paths, we can store them using `StoreDerivedAccounts(id, passwordToEncryptKey, paths)`.
`StoreDerivedAccounts` returns an address/pubKey for each path. The address can be use in the future to load them in memory again.
Calling `StoreDerivedAccounts` will encrypt and store the keys, each one in a keystore json file, and remove all the keys from memory.
Since they are derived from an extended key, they are extended keys too, so they can be used in the future to derive more child keys.
`StoreAccount` stores the key identified by its ID, so in case the key comes from `Generate` or `ImportPrivateKey`, it will store the master key.
In general we want to avoid saving master keys, so we should only use `StoreDerivedAccounts` for extended keys, and `StoreAccount` for normal keys.

Calling `Load(address, password)` will unlock the key specified by addresses using password, and load it in memory.
`Load` returns a new id that can be used again with DeriveAddresses, `StoreAccount`, and `StoreDerivedAccounts`.

`ImportPrivateKey` imports a raw private key specified by its hex form.
It's not an extended key, so it can't be used to derive child addresses.
You can call `DeriveAddresses` to derive the address/pubKey of a normal key passing an empty string as derivation path.
`StoreAccount` will save the key without deriving a child key.




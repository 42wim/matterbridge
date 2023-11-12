#!/bin/sh -e

if [ $# -ne 1 ]
then
  echo "Usage: $0 libtomcrypt_dir" >&2
  echo "Copy tracked source files from libtomcrypt_dir to current directory." >&2
  exit 1
fi

ltd=$1

# copy header files
cp -f $ltd/src/headers/tomcrypt_argchk.h .
cp -f $ltd/src/headers/tomcrypt_cfg.h .
cp -f $ltd/src/headers/tomcrypt_cipher.h .
cp -f $ltd/src/headers/tomcrypt_custom.h .
cp -f $ltd/src/headers/tomcrypt.h .
cp -f $ltd/src/headers/tomcrypt_hash.h .
cp -f $ltd/src/headers/tomcrypt_mac.h .
cp -f $ltd/src/headers/tomcrypt_macros.h .
cp -f $ltd/src/headers/tomcrypt_math.h .
cp -f $ltd/src/headers/tomcrypt_misc.h .
cp -f $ltd/src/headers/tomcrypt_pkcs.h .
cp -f $ltd/src/headers/tomcrypt_pk.h .
cp -f $ltd/src/headers/tomcrypt_private.h .
cp -f $ltd/src/headers/tomcrypt_prng.h .

# copy C files
cp -f $ltd/src/ciphers/aes/aes.c .
cp -f $ltd/src/ciphers/aes/aes_tab.c aes_tab.h
cp -f $ltd/src/misc/burn_stack.c .
cp -f $ltd/src/misc/compare_testvector.c .
cp -f $ltd/src/modes/cbc/cbc_decrypt.c .
cp -f $ltd/src/modes/cbc/cbc_done.c .
cp -f $ltd/src/modes/cbc/cbc_encrypt.c .
cp -f $ltd/src/modes/cbc/cbc_start.c .
cp -f $ltd/src/misc/crypt/crypt_argchk.c .
cp -f $ltd/src/misc/crypt/crypt_cipher_descriptor.c .
cp -f $ltd/src/misc/crypt/crypt_cipher_is_valid.c .
cp -f $ltd/src/misc/crypt/crypt_find_cipher.c .
cp -f $ltd/src/misc/crypt/crypt_find_hash.c .
cp -f $ltd/src/misc/crypt/crypt_hash_descriptor.c .
cp -f $ltd/src/misc/crypt/crypt_hash_is_valid.c .
cp -f $ltd/src/misc/crypt/crypt_prng_descriptor.c .
cp -f $ltd/src/misc/crypt/crypt_register_cipher.c .
cp -f $ltd/src/misc/crypt/crypt_register_hash.c .
cp -f $ltd/src/misc/crypt/crypt_register_prng.c .
cp -f $ltd/src/prngs/fortuna.c .
cp -f $ltd/src/hashes/helper/hash_memory.c .
cp -f $ltd/src/mac/hmac/hmac_done.c .
cp -f $ltd/src/mac/hmac/hmac_init.c .
cp -f $ltd/src/mac/hmac/hmac_memory.c .
cp -f $ltd/src/mac/hmac/hmac_process.c .
cp -f $ltd/src/misc/pkcs5/pkcs_5_2.c .
cp -f $ltd/src/hashes/sha1.c .
cp -f $ltd/src/hashes/sha2/sha256.c .
cp -f $ltd/src/hashes/sha2/sha512.c .
cp -f $ltd/src/misc/zeromem.c .

echo "make sure aes.c includes aes_tab.h instead of aes_tab.c!"

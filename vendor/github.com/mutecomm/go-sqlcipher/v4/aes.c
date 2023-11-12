/* LibTomCrypt, modular cryptographic library -- Tom St Denis
 *
 * LibTomCrypt is a library that provides various cryptographic
 * algorithms in a highly modular and flexible manner.
 *
 * The library is free for all purposes without any express
 * guarantee it works.
 *
 * Tom St Denis, tomstdenis@gmail.com, http://libtom.org
 */

/**
  @file aes.c
  Implementation of AES
*/

#include "tomcrypt_private.h"

#ifdef LTC_RIJNDAEL

#include "mbtls_aes.h"

static __thread mbedtls_aes_context ctx_encrypt;

#ifndef ENCRYPT_ONLY
static __thread mbedtls_aes_context ctx_decrypt;

#define SETUP    rijndael_setup
#define ECB_ENC  rijndael_ecb_encrypt
#define ECB_DEC  rijndael_ecb_decrypt
#define ECB_DONE rijndael_done
#define ECB_TEST rijndael_test
#define ECB_KS   rijndael_keysize

const struct ltc_cipher_descriptor rijndael_desc =
{
    "rijndael",
    6,
    16, 32, 16, 10,
    SETUP, ECB_ENC, ECB_DEC, ECB_TEST, ECB_DONE, ECB_KS,
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
};

const struct ltc_cipher_descriptor aes_desc =
{
    "aes",
    6,
    16, 32, 16, 10,
    SETUP, ECB_ENC, ECB_DEC, ECB_TEST, ECB_DONE, ECB_KS,
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
};

#else

#define SETUP    rijndael_enc_setup
#define ECB_ENC  rijndael_enc_ecb_encrypt
#define ECB_KS   rijndael_enc_keysize
#define ECB_DONE rijndael_enc_done

const struct ltc_cipher_descriptor rijndael_enc_desc =
{
    "rijndael",
    6,
    16, 32, 16, 10,
    SETUP, ECB_ENC, NULL, NULL, ECB_DONE, ECB_KS,
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
};

const struct ltc_cipher_descriptor aes_enc_desc =
{
    "aes",
    6,
    16, 32, 16, 10,
    SETUP, ECB_ENC, NULL, NULL, ECB_DONE, ECB_KS,
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
};

#endif

 /**
    Initialize the AES (Rijndael) block cipher
    @param key The symmetric key you wish to pass
    @param keylen The key length in bytes
    @param num_rounds The number of rounds desired (0 for default)
    @param skey The key in as scheduled by this function.
    @return CRYPT_OK if successful
 */
int SETUP(const unsigned char *key, int keylen, int num_rounds, symmetric_key *skey)
{
    LTC_ARGCHK(key  != NULL);
    LTC_ARGCHK(skey != NULL);

    if (keylen != 16 && keylen != 24 && keylen != 32) {
       return CRYPT_INVALID_KEYSIZE;
    }

    if (num_rounds != 0 && num_rounds != (10 + ((keylen/8)-2)*2)) {
       return CRYPT_INVALID_ROUNDS;
    }

    mbedtls_aes_init(&ctx_encrypt);
    if (mbedtls_aes_setkey_enc(&ctx_encrypt, key, keylen*8) != 0)
       return CRYPT_INVALID_KEYSIZE;
    memcpy(skey->rijndael.eK, ctx_encrypt.buf, sizeof(skey->rijndael.eK));

#ifndef ENCRYPT_ONLY
    mbedtls_aes_init(&ctx_decrypt);
    if (mbedtls_aes_setkey_dec(&ctx_decrypt, key, keylen*8) != 0)
       return CRYPT_INVALID_KEYSIZE;
    memcpy(skey->rijndael.dK, ctx_decrypt.buf, sizeof(skey->rijndael.dK));
#endif

    skey->rijndael.Nr = ctx_encrypt.nr;

    return CRYPT_OK;
}

/**
  Encrypts a block of text with AES
  @param pt The input plaintext (16 bytes)
  @param ct The output ciphertext (16 bytes)
  @param skey The key as scheduled
  @return CRYPT_OK if successful
*/
#ifdef LTC_CLEAN_STACK
static int s_rijndael_ecb_encrypt(const unsigned char *pt, unsigned char *ct, const symmetric_key *skey)
#else
int ECB_ENC(const unsigned char *pt, unsigned char *ct, const symmetric_key *skey)
#endif
{
    LTC_ARGCHK(pt != NULL);
    LTC_ARGCHK(ct != NULL);
    LTC_ARGCHK(skey != NULL);

    ctx_encrypt.nr = skey->rijndael.Nr;
    memset(ctx_encrypt.buf, 0, sizeof(ctx_encrypt.buf));
    memcpy(ctx_encrypt.buf, skey->rijndael.eK, sizeof(skey->rijndael.eK));

    return mbedtls_aes_crypt_ecb(&ctx_encrypt, MBEDTLS_AES_ENCRYPT, pt, ct) == 0 ? CRYPT_OK : CRYPT_ERROR;
}

#ifdef LTC_CLEAN_STACK
int ECB_ENC(const unsigned char *pt, unsigned char *ct, const symmetric_key *skey)
{
    return s_rijndael_ecb_encrypt(pt, ct, skey);
}
#endif

#ifndef ENCRYPT_ONLY

/**
  Decrypts a block of text with AES
  @param ct The input ciphertext (16 bytes)
  @param pt The output plaintext (16 bytes)
  @param skey The key as scheduled
  @return CRYPT_OK if successful
*/
#ifdef LTC_CLEAN_STACK
static int s_rijndael_ecb_decrypt(const unsigned char *ct, unsigned char *pt, const symmetric_key *skey)
#else
int ECB_DEC(const unsigned char *ct, unsigned char *pt, const symmetric_key *skey)
#endif
{
    LTC_ARGCHK(pt != NULL);
    LTC_ARGCHK(ct != NULL);
    LTC_ARGCHK(skey != NULL);

    ctx_decrypt.nr = skey->rijndael.Nr;
    memset(ctx_decrypt.buf, 0, sizeof(ctx_decrypt.buf));
    memcpy(ctx_decrypt.buf, skey->rijndael.dK, sizeof(skey->rijndael.dK));

    return mbedtls_aes_crypt_ecb(&ctx_decrypt, MBEDTLS_AES_DECRYPT, ct, pt) == 0 ? CRYPT_OK : CRYPT_ERROR;
}


#ifdef LTC_CLEAN_STACK
int ECB_DEC(const unsigned char *ct, unsigned char *pt, const symmetric_key *skey)
{
    return s_rijndael_ecb_decrypt(ct, pt, skey);
}
#endif

/**
  Performs a self-test of the AES block cipher
  @return CRYPT_OK if functional, CRYPT_NOP if self-test has been disabled
*/
int ECB_TEST(void)
{
 #ifndef LTC_TEST
    return CRYPT_NOP;
 #else
 int err;
 static const struct {
     int keylen;
     unsigned char key[32], pt[16], ct[16];
 } tests[] = {
    { 16,
      { 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
        0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f },
      { 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
        0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff },
      { 0x69, 0xc4, 0xe0, 0xd8, 0x6a, 0x7b, 0x04, 0x30,
        0xd8, 0xcd, 0xb7, 0x80, 0x70, 0xb4, 0xc5, 0x5a }
    }, {
      24,
      { 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
        0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
        0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17 },
      { 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
        0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff },
      { 0xdd, 0xa9, 0x7c, 0xa4, 0x86, 0x4c, 0xdf, 0xe0,
        0x6e, 0xaf, 0x70, 0xa0, 0xec, 0x0d, 0x71, 0x91 }
    }, {
      32,
      { 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
        0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
        0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
        0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f },
      { 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
        0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff },
      { 0x8e, 0xa2, 0xb7, 0xca, 0x51, 0x67, 0x45, 0xbf,
        0xea, 0xfc, 0x49, 0x90, 0x4b, 0x49, 0x60, 0x89 }
    }
 };

  symmetric_key key;
  unsigned char tmp[2][16];
  int i, y;

  for (i = 0; i < (int)(sizeof(tests)/sizeof(tests[0])); i++) {
    zeromem(&key, sizeof(key));
    if ((err = rijndael_setup(tests[i].key, tests[i].keylen, 0, &key)) != CRYPT_OK) {
       return err;
    }

    rijndael_ecb_encrypt(tests[i].pt, tmp[0], &key);
    rijndael_ecb_decrypt(tmp[0], tmp[1], &key);
    if (compare_testvector(tmp[0], 16, tests[i].ct, 16, "AES Encrypt", i) ||
          compare_testvector(tmp[1], 16, tests[i].pt, 16, "AES Decrypt", i)) {
        return CRYPT_FAIL_TESTVECTOR;
    }

    /* now see if we can encrypt all zero bytes 1000 times, decrypt and come back where we started */
    for (y = 0; y < 16; y++) tmp[0][y] = 0;
    for (y = 0; y < 1000; y++) rijndael_ecb_encrypt(tmp[0], tmp[0], &key);
    for (y = 0; y < 1000; y++) rijndael_ecb_decrypt(tmp[0], tmp[0], &key);
    for (y = 0; y < 16; y++) if (tmp[0][y] != 0) return CRYPT_FAIL_TESTVECTOR;
  }
  return CRYPT_OK;
 #endif
}

#endif /* ENCRYPT_ONLY */


/** Terminate the context
   @param skey    The scheduled key
*/
void ECB_DONE(symmetric_key *skey)
{
    mbedtls_aes_free(&ctx_encrypt);
#ifndef ENCRYPT_ONLY
    mbedtls_aes_free(&ctx_decrypt);
#endif
}


/**
  Gets suitable key size
  @param keysize [in/out] The length of the recommended key (in bytes).  This function will store the suitable size back in this variable.
  @return CRYPT_OK if the input key size is acceptable.
*/
int ECB_KS(int *keysize)
{
   LTC_ARGCHK(keysize != NULL);

   if (*keysize < 16) {
      return CRYPT_INVALID_KEYSIZE;
   }
   if (*keysize < 24) {
      *keysize = 16;
      return CRYPT_OK;
   }
   if (*keysize < 32) {
      *keysize = 24;
      return CRYPT_OK;
   }
   *keysize = 32;
   return CRYPT_OK;
}

#endif

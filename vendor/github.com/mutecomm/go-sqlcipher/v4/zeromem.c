/* LibTomCrypt, modular cryptographic library -- Tom St Denis */
/* SPDX-License-Identifier: Unlicense */
#include "tomcrypt_private.h"
#include <string.h>

/**
   @file zeromem.c
   Zero a block of memory, Tom St Denis
*/

/**
   Zero a block of memory
   @param out    The destination of the area to zero
   @param outlen The length of the area to zero (octets)
*/
void zeromem(volatile void *out, size_t outlen)
{
   LTC_ARGCHKVD(out != NULL);
   memset((void *)out, 0, outlen);
}

/* $Source$ */
/* $Revision$ */
/* $Date$ */

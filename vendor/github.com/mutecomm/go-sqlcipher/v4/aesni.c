/*
 *  AES-NI support functions
 *
 *  Copyright The Mbed TLS Contributors
 *  SPDX-License-Identifier: Apache-2.0
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may
 *  not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 *  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * [AES-WP] https://www.intel.com/content/www/us/en/developer/articles/tool/intel-advanced-encryption-standard-aes-instructions-set.html
 * [CLMUL-WP] https://www.intel.com/content/www/us/en/develop/download/intel-carry-less-multiplication-instruction-and-its-usage-for-computing-the-gcm-mode.html
 */

#include "aesni.h"

#if defined(MBEDTLS_AESNI_C)

#include <string.h>

#if defined(MBEDTLS_AESNI_HAVE_CODE)

/*
 * AES-NI support detection routine
 */
int mbedtls_aesni_has_support(unsigned int what)
{
    static int done = 0;
    static unsigned int c = 0;

    if (!done) {
        /* AESNI using asm */
        asm ("movl  $1, %%eax   \n\t"
             "cpuid             \n\t"
             : "=c" (c)
             :
             : "eax", "ebx", "edx");
        done = 1;
    }

    return (c & what) != 0;
}

/*
 * Binutils needs to be at least 2.19 to support AES-NI instructions.
 * Unfortunately, a lot of users have a lower version now (2014-04).
 * Emit bytecode directly in order to support "old" version of gas.
 *
 * Opcodes from the Intel architecture reference manual, vol. 3.
 * We always use registers, so we don't need prefixes for memory operands.
 * Operand macros are in gas order (src, dst) as opposed to Intel order
 * (dst, src) in order to blend better into the surrounding assembly code.
 */
#define AESDEC(regs)      ".byte 0x66,0x0F,0x38,0xDE," regs "\n\t"
#define AESDECLAST(regs)  ".byte 0x66,0x0F,0x38,0xDF," regs "\n\t"
#define AESENC(regs)      ".byte 0x66,0x0F,0x38,0xDC," regs "\n\t"
#define AESENCLAST(regs)  ".byte 0x66,0x0F,0x38,0xDD," regs "\n\t"
#define AESIMC(regs)      ".byte 0x66,0x0F,0x38,0xDB," regs "\n\t"
#define AESKEYGENA(regs, imm)  ".byte 0x66,0x0F,0x3A,0xDF," regs "," imm "\n\t"
#define PCLMULQDQ(regs, imm)   ".byte 0x66,0x0F,0x3A,0x44," regs "," imm "\n\t"

#define xmm0_xmm0   "0xC0"
#define xmm0_xmm1   "0xC8"
#define xmm0_xmm2   "0xD0"
#define xmm0_xmm3   "0xD8"
#define xmm0_xmm4   "0xE0"
#define xmm1_xmm0   "0xC1"
#define xmm1_xmm2   "0xD1"

/*
 * AES-NI AES-ECB block en(de)cryption
 */
int mbedtls_aesni_crypt_ecb(mbedtls_aes_context *ctx, int mode, const unsigned char input[16], unsigned char output[16])
{
    asm ("movdqu    (%3), %%xmm0    \n\t" // load input
         "movdqu    (%1), %%xmm1    \n\t" // load round key 0
         "pxor      %%xmm1, %%xmm0  \n\t" // round 0
         "add       $16, %1         \n\t" // point to next round key
         "subl      $1, %0          \n\t" // normal rounds = nr - 1
         "test      %2, %2          \n\t" // mode?
         "jz        2f              \n\t" // 0 = decrypt

         "1:                        \n\t" // encryption loop
         "movdqu    (%1), %%xmm1    \n\t" // load round key
         AESENC(xmm1_xmm0)                // do round
         "add       $16, %1         \n\t" // point to next round key
         "subl      $1, %0          \n\t" // loop
         "jnz       1b              \n\t"
         "movdqu    (%1), %%xmm1    \n\t" // load round key
         AESENCLAST(xmm1_xmm0)            // last round
         "jmp       3f              \n\t"

         "2:                        \n\t" // decryption loop
         "movdqu    (%1), %%xmm1    \n\t"
         AESDEC(xmm1_xmm0)                // do round
         "add       $16, %1         \n\t"
         "subl      $1, %0          \n\t"
         "jnz       2b              \n\t"
         "movdqu    (%1), %%xmm1    \n\t" // load round key
         AESDECLAST(xmm1_xmm0)            // last round

         "3:                        \n\t"
         "movdqu    %%xmm0, (%4)    \n\t" // export output
         :
         : "r" (ctx->nr), "r" (ctx->buf + ctx->rk_offset), "r" (mode), "r" (input), "r" (output)
         : "memory", "cc", "xmm0", "xmm1");


    return 0;
}

/*
 * Compute decryption round keys from encryption round keys
 */
void mbedtls_aesni_inverse_key(unsigned char *invkey, const unsigned char *fwdkey, int nr)
{
    unsigned char *ik = invkey;
    const unsigned char *fk = fwdkey + 16 * nr;

    memcpy(ik, fk, 16);

    for (fk -= 16, ik += 16; fk > fwdkey; fk -= 16, ik += 16) {
        asm ("movdqu (%0), %%xmm0       \n\t"
             AESIMC(xmm0_xmm0)
             "movdqu %%xmm0, (%1)       \n\t"
             :
             : "r" (fk), "r" (ik)
             : "memory", "xmm0");
    }

    memcpy(ik, fk, 16);
}

/*
 * Key expansion, 128-bit case
 */
static void aesni_setkey_enc_128(unsigned char *rk, const unsigned char *key)
{
    asm ("movdqu (%1), %%xmm0               \n\t" // copy the original key
         "movdqu %%xmm0, (%0)               \n\t" // as round key 0
         "jmp 2f                            \n\t" // skip auxiliary routine

         /*
          * Finish generating the next round key.
          *
          * On entry xmm0 is r3:r2:r1:r0 and xmm1 is X:stuff:stuff:stuff
          * with X = rot( sub( r3 ) ) ^ RCON.
          *
          * On exit, xmm0 is r7:r6:r5:r4
          * with r4 = X + r0, r5 = r4 + r1, r6 = r5 + r2, r7 = r6 + r3
          * and those are written to the round key buffer.
          */
         "1:                                \n\t"
         "pshufd $0xff, %%xmm1, %%xmm1      \n\t" // X:X:X:X
         "pxor %%xmm0, %%xmm1               \n\t" // X+r3:X+r2:X+r1:r4
         "pslldq $4, %%xmm0                 \n\t" // r2:r1:r0:0
         "pxor %%xmm0, %%xmm1               \n\t" // X+r3+r2:X+r2+r1:r5:r4
         "pslldq $4, %%xmm0                 \n\t" // etc
         "pxor %%xmm0, %%xmm1               \n\t"
         "pslldq $4, %%xmm0                 \n\t"
         "pxor %%xmm1, %%xmm0               \n\t" // update xmm0 for next time!
         "add $16, %0                       \n\t" // point to next round key
         "movdqu %%xmm0, (%0)               \n\t" // write it
         "ret                               \n\t"

         /* Main "loop" */
         "2:                                \n\t"
         AESKEYGENA(xmm0_xmm1, "0x01")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x02")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x04")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x08")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x10")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x20")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x40")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x80")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x1B")      "call 1b \n\t"
         AESKEYGENA(xmm0_xmm1, "0x36")      "call 1b \n\t"
         :
         : "r" (rk), "r" (key)
         : "memory", "cc", "0");
}

/*
 * Key expansion, 192-bit case
 */
static void aesni_setkey_enc_192(unsigned char *rk, const unsigned char *key)
{
    asm ("movdqu (%1), %%xmm0   \n\t" // copy original round key
         "movdqu %%xmm0, (%0)   \n\t"
         "add $16, %0           \n\t"
         "movq 16(%1), %%xmm1   \n\t"
         "movq %%xmm1, (%0)     \n\t"
         "add $8, %0            \n\t"
         "jmp 2f                \n\t" // skip auxiliary routine

         /*
          * Finish generating the next 6 quarter-keys.
          *
          * On entry xmm0 is r3:r2:r1:r0, xmm1 is stuff:stuff:r5:r4
          * and xmm2 is stuff:stuff:X:stuff with X = rot( sub( r3 ) ) ^ RCON.
          *
          * On exit, xmm0 is r9:r8:r7:r6 and xmm1 is stuff:stuff:r11:r10
          * and those are written to the round key buffer.
          */
         "1:                            \n\t"
         "pshufd $0x55, %%xmm2, %%xmm2  \n\t" // X:X:X:X
         "pxor %%xmm0, %%xmm2           \n\t" // X+r3:X+r2:X+r1:r4
         "pslldq $4, %%xmm0             \n\t" // etc
         "pxor %%xmm0, %%xmm2           \n\t"
         "pslldq $4, %%xmm0             \n\t"
         "pxor %%xmm0, %%xmm2           \n\t"
         "pslldq $4, %%xmm0             \n\t"
         "pxor %%xmm2, %%xmm0           \n\t" // update xmm0 = r9:r8:r7:r6
         "movdqu %%xmm0, (%0)           \n\t"
         "add $16, %0                   \n\t"
         "pshufd $0xff, %%xmm0, %%xmm2  \n\t" // r9:r9:r9:r9
         "pxor %%xmm1, %%xmm2           \n\t" // stuff:stuff:r9+r5:r10
         "pslldq $4, %%xmm1             \n\t" // r2:r1:r0:0
         "pxor %%xmm2, %%xmm1           \n\t" // xmm1 = stuff:stuff:r11:r10
         "movq %%xmm1, (%0)             \n\t"
         "add $8, %0                    \n\t"
         "ret                           \n\t"

         "2:                            \n\t"
         AESKEYGENA(xmm1_xmm2, "0x01")  "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x02")  "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x04")  "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x08")  "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x10")  "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x20")  "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x40")  "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x80")  "call 1b \n\t"

         :
         : "r" (rk), "r" (key)
         : "memory", "cc", "0");
}

/*
 * Key expansion, 256-bit case
 */
static void aesni_setkey_enc_256(unsigned char *rk, const unsigned char *key)
{
    asm ("movdqu (%1), %%xmm0           \n\t"
         "movdqu %%xmm0, (%0)           \n\t"
         "add $16, %0                   \n\t"
         "movdqu 16(%1), %%xmm1         \n\t"
         "movdqu %%xmm1, (%0)           \n\t"
         "jmp 2f                        \n\t" // skip auxiliary routine

         /*
          * Finish generating the next two round keys.
          *
          * On entry xmm0 is r3:r2:r1:r0, xmm1 is r7:r6:r5:r4 and
          * xmm2 is X:stuff:stuff:stuff with X = rot( sub( r7 )) ^ RCON
          *
          * On exit, xmm0 is r11:r10:r9:r8 and xmm1 is r15:r14:r13:r12
          * and those have been written to the output buffer.
          */
         "1:                                \n\t"
         "pshufd $0xff, %%xmm2, %%xmm2      \n\t"
         "pxor %%xmm0, %%xmm2               \n\t"
         "pslldq $4, %%xmm0                 \n\t"
         "pxor %%xmm0, %%xmm2               \n\t"
         "pslldq $4, %%xmm0                 \n\t"
         "pxor %%xmm0, %%xmm2               \n\t"
         "pslldq $4, %%xmm0                 \n\t"
         "pxor %%xmm2, %%xmm0               \n\t"
         "add $16, %0                       \n\t"
         "movdqu %%xmm0, (%0)               \n\t"

         /* Set xmm2 to stuff:Y:stuff:stuff with Y = subword( r11 )
          * and proceed to generate next round key from there */
         AESKEYGENA(xmm0_xmm2, "0x00")
         "pshufd $0xaa, %%xmm2, %%xmm2      \n\t"
         "pxor %%xmm1, %%xmm2               \n\t"
         "pslldq $4, %%xmm1                 \n\t"
         "pxor %%xmm1, %%xmm2               \n\t"
         "pslldq $4, %%xmm1                 \n\t"
         "pxor %%xmm1, %%xmm2               \n\t"
         "pslldq $4, %%xmm1                 \n\t"
         "pxor %%xmm2, %%xmm1               \n\t"
         "add $16, %0                       \n\t"
         "movdqu %%xmm1, (%0)               \n\t"
         "ret                               \n\t"

         /*
          * Main "loop" - Generating one more key than necessary,
          * see definition of mbedtls_aes_context.buf
          */
         "2:                                \n\t"
         AESKEYGENA(xmm1_xmm2, "0x01")      "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x02")      "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x04")      "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x08")      "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x10")      "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x20")      "call 1b \n\t"
         AESKEYGENA(xmm1_xmm2, "0x40")      "call 1b \n\t"
         :
         : "r" (rk), "r" (key)
         : "memory", "cc", "0");
}

/*
 * Key expansion, wrapper
 */
int mbedtls_aesni_setkey_enc(unsigned char *rk, const unsigned char *key, size_t bits)
{
    switch (bits) {
        case 128: aesni_setkey_enc_128(rk, key); break;
        case 192: aesni_setkey_enc_192(rk, key); break;
        case 256: aesni_setkey_enc_256(rk, key); break;
        default: return MBEDTLS_ERR_AES_INVALID_KEY_LENGTH;
    }

    return 0;
}

#endif /* MBEDTLS_AESNI_HAVE_CODE */

#endif /* MBEDTLS_AESNI_C */

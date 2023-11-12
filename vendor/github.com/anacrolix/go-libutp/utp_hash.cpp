/*
 * Copyright (c) 2010-2013 BitTorrent, Inc.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

#include "utp_hash.h"
#include "utp_types.h"

#define LIBUTP_HASH_UNUSED ((utp_link_t)-1)

#ifdef STRICT_ALIGN
inline uint32 Read32(const void *p)
{
	uint32 tmp;
	memcpy(&tmp, p, sizeof tmp);
	return tmp;
}

#else
inline uint32 Read32(const void *p) { return *(uint32*)p; }
#endif


// Get the amount of memory required for the hash parameters and the bucket set
// Waste a space for an unused bucket in order to ensure the following managed memory have 32-bit aligned addresses
// TODO:  make this 64-bit clean
#define BASE_SIZE(bc) (sizeof(utp_hash_t) + sizeof(utp_link_t) * ((bc) + 1))

// Get a pointer to the base of the structure array managed by the hash table
#define get_bep(h) ((byte*)(h)) + BASE_SIZE((h)->N)

// Get the address of the information associated with a specific structure in the array,
// given the address of the base of the structure.
// This assumes a utp_link_t link member is at the end of the structure.
// Given compilers filling out the memory to a 32-bit clean value, this may mean that
// the location named in the structure may not be the location actually used by the hash table,
// since the compiler may have padded the end of the structure with 2 bytes after the utp_link_t member.
// TODO: this macro should not require that the variable pointing at the hash table be named 'hash'
#define ptr_to_link(p) (utp_link_t *) (((byte *) (p)) + hash->E - sizeof(utp_link_t))

// Calculate how much to allocate for a hash table with bucket count, total size, and structure count
// TODO:  make this 64-bit clean
#define ALLOCATION_SIZE(bc, ts, sc) (BASE_SIZE((bc)) + (ts) * (sc))

utp_hash_t *utp_hash_create(int N, int key_size, int total_size, int initial, utp_hash_compute_t hashfun, utp_hash_equal_t compfun)
{
	// Must have odd number of hash buckets (prime number is best)
	assert(N % 2);
	// Ensure structures will be at aligned memory addresses
	// TODO:  make this 64-bit clean
	assert(0 == (total_size % 4));

	int size = ALLOCATION_SIZE(N, total_size, initial);
	utp_hash_t *hash = (utp_hash_t *) malloc( size );
	memset( hash, 0, size );

	for (int i = 0; i < N + 1; ++i)
		hash->inits[i] = LIBUTP_HASH_UNUSED;
	hash->N = N;
	hash->K = key_size;
	hash->E = total_size;
	hash->hash_compute = hashfun;
	hash->hash_equal = compfun;
	hash->allocated = initial;
	hash->count = 0;
	hash->used = 0;
	hash->free = LIBUTP_HASH_UNUSED;
	return hash;
}

uint utp_hash_mem(const void *keyp, size_t keysize)
{
	uint hash = 0;
	uint n = keysize;
	while (n >= 4) {
		hash ^= Read32(keyp);
		keyp = (byte*)keyp + sizeof(uint32);
		hash = (hash << 13) | (hash >> 19);
		n -= 4;
	}
	while (n != 0) {
		hash ^= *(byte*)keyp;
		keyp = (byte*)keyp + sizeof(byte);
		hash = (hash << 8) | (hash >> 24);
		n--;
	}
	return hash;
}

uint utp_hash_mkidx(utp_hash_t *hash, const void *keyp)
{
	// Generate a key from the hash
	return hash->hash_compute(keyp, hash->K) % hash->N;
}

static inline bool compare(byte *a, byte *b,int n)
{
	assert(n >= 4);
	if (Read32(a) != Read32(b)) return false;
	return memcmp(a+4, b+4, n-4) == 0;
}

#define COMPARE(h,k1,k2,ks) (((h)->hash_equal) ? (h)->hash_equal((void*)k1,(void*)k2,ks) : compare(k1,k2,ks))

// Look-up a key in the hash table.
// Returns NULL if not found
void *utp_hash_lookup(utp_hash_t *hash, const void *key)
{
	utp_link_t idx = utp_hash_mkidx(hash, key);

	// base pointer
	byte *bep = get_bep(hash);

	utp_link_t cur = hash->inits[idx];
	while (cur != LIBUTP_HASH_UNUSED) {
		byte *key2 = bep + (cur * hash->E);
		if (COMPARE(hash, (byte*)key, key2, hash->K))
			return key2;
		cur = *ptr_to_link(key2);
	}

	return NULL;
}

// Add a new element to the hash table.
// Returns a pointer to the new element.
// This assumes the element is not already present!
void *utp_hash_add(utp_hash_t **hashp, const void *key)
{
	//Allocate a new entry
	byte *elemp;
	utp_link_t elem;
	utp_hash_t *hash = *hashp;
	utp_link_t idx = utp_hash_mkidx(hash, key);

	if ((elem=hash->free) == LIBUTP_HASH_UNUSED) {
		utp_link_t all = hash->allocated;
		if (hash->used == all) {
			utp_hash_t *nhash;
			if (all <= (LIBUTP_HASH_UNUSED/2)) {
				all *= 2;
			} else if (all != LIBUTP_HASH_UNUSED) {
				all  = LIBUTP_HASH_UNUSED;
			} else {
				// too many items! can't grow!
				assert(0);
				return NULL;
			}
			// otherwise need to allocate.
			nhash = (utp_hash_t*)realloc(hash, ALLOCATION_SIZE(hash->N, hash->E, all));
			if (!nhash) {
				// out of memory (or too big to allocate)
				assert(nhash);
				return NULL;
			}
			hash = *hashp = nhash;
			hash->allocated = all;
		}

		elem = hash->used++;
		elemp = get_bep(hash) + elem * hash->E;
	} else {
		elemp = get_bep(hash) + elem * hash->E;
		hash->free = *ptr_to_link(elemp);
	}

	*ptr_to_link(elemp) = hash->inits[idx];
	hash->inits[idx] = elem;
	hash->count++;

	// copy key into it
	memcpy(elemp, key, hash->K);
	return elemp;
}

// Delete an element from the utp_hash_t
// Returns a pointer to the already deleted element.
void *utp_hash_del(utp_hash_t *hash, const void *key)
{
	utp_link_t idx = utp_hash_mkidx(hash, key);

	// base pointer
	byte *bep = get_bep(hash);

	utp_link_t *curp = &hash->inits[idx];
	utp_link_t cur;
	while ((cur=*curp) != LIBUTP_HASH_UNUSED) {
		byte *key2 = bep + (cur * hash->E);
		if (COMPARE(hash,(byte*)key,(byte*)key2, hash->K )) {
			// found an item that matched. unlink it
			*curp = *ptr_to_link(key2);
			// Insert into freelist
			*ptr_to_link(key2) = hash->free;
			hash->free = cur;
			hash->count--;
			return key2;
		}
		curp = ptr_to_link(key2);
	}

	return NULL;
}

void *utp_hash_iterate(utp_hash_t *hash, utp_hash_iterator_t *iter)
{
	utp_link_t elem;

	if ((elem=iter->elem) == LIBUTP_HASH_UNUSED) {
		// Find a bucket with an element
		utp_link_t buck = iter->bucket + 1;
		for(;;) {
			if (buck >= hash->N)
				return NULL;
			if ((elem = hash->inits[buck]) != LIBUTP_HASH_UNUSED)
				break;
			buck++;
		}
		iter->bucket = buck;
	}

	byte *elemp = get_bep(hash) + (elem * hash->E);
	iter->elem = *ptr_to_link(elemp);
	return elemp;
}

void utp_hash_free_mem(utp_hash_t* hash)
{
	free(hash);
}

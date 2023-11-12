package dnscache

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type DNSResolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	LookupAddr(ctx context.Context, addr string) (names []string, err error)
}

type Resolver struct {
	// Timeout defines the maximum allowed time allowed for a lookup.
	Timeout time.Duration

	// Resolver is used to perform actual DNS lookup. If nil,
	// net.DefaultResolver is used instead.
	Resolver DNSResolver

	once  sync.Once
	mu    sync.RWMutex
	cache map[string]*cacheEntry

	// OnCacheMiss is executed if the host or address is not included in
	// the cache and the default lookup is executed.
	OnCacheMiss func()
}

type ResolverRefreshOptions struct {
	ClearUnused      bool
	PersistOnFailure bool
}

type cacheEntry struct {
	rrs  []string
	err  error
	used bool
}

// LookupAddr performs a reverse lookup for the given address, returning a list
// of names mapping to that address.
func (r *Resolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	r.once.Do(r.init)
	return r.lookup(ctx, "r"+addr)
}

// LookupHost looks up the given host using the local resolver. It returns a
// slice of that host's addresses.
func (r *Resolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	r.once.Do(r.init)
	return r.lookup(ctx, "h"+host)
}

// refreshRecords refreshes cached entries which have been used at least once since
// the last Refresh. If clearUnused is true, entries which haven't be used since the
// last Refresh are removed from the cache. If persistOnFailure is true, stale
// entries will not be removed on failed lookups
func (r *Resolver) refreshRecords(clearUnused bool, persistOnFailure bool) {
	r.once.Do(r.init)
	r.mu.RLock()
	update := make([]string, 0, len(r.cache))
	del := make([]string, 0, len(r.cache))
	for key, entry := range r.cache {
		if entry.used {
			update = append(update, key)
		} else if clearUnused {
			del = append(del, key)
		}
	}
	r.mu.RUnlock()

	if len(del) > 0 {
		r.mu.Lock()
		for _, key := range del {
			delete(r.cache, key)
		}
		r.mu.Unlock()
	}

	for _, key := range update {
		r.update(context.Background(), key, false, persistOnFailure)
	}
}

func (r *Resolver) Refresh(clearUnused bool) {
	r.refreshRecords(clearUnused, false)
}

func (r *Resolver) RefreshWithOptions(options ResolverRefreshOptions) {
	r.refreshRecords(options.ClearUnused, options.PersistOnFailure)
}

func (r *Resolver) init() {
	r.cache = make(map[string]*cacheEntry)
}

// lookupGroup merges lookup calls together for lookups for the same host. The
// lookupGroup key is is the LookupIPAddr.host argument.
var lookupGroup singleflight.Group

func (r *Resolver) lookup(ctx context.Context, key string) (rrs []string, err error) {
	var found bool
	rrs, err, found = r.load(key)
	if !found {
		if r.OnCacheMiss != nil {
			r.OnCacheMiss()
		}
		rrs, err = r.update(ctx, key, true, false)
	}
	return
}

func (r *Resolver) update(ctx context.Context, key string, used bool, persistOnFailure bool) (rrs []string, err error) {
	c := lookupGroup.DoChan(key, r.lookupFunc(key))
	select {
	case <-ctx.Done():
		err = ctx.Err()
		if err == context.DeadlineExceeded {
			// If DNS request timed out for some reason, force future
			// request to start the DNS lookup again rather than waiting
			// for the current lookup to complete.
			lookupGroup.Forget(key)
		}
	case res := <-c:
		if res.Shared {
			// We had concurrent lookups, check if the cache is already updated
			// by a friend.
			var found bool
			rrs, err, found = r.load(key)
			if found {
				return
			}
		}
		err = res.Err
		if err == nil {
			rrs, _ = res.Val.([]string)
		}

		if err != nil && persistOnFailure {
			var found bool
			rrs, err, found = r.load(key)
			if found {
				return
			}
		}

		r.mu.Lock()
		r.storeLocked(key, rrs, used, err)
		r.mu.Unlock()
	}
	return
}

// lookupFunc returns lookup function for key. The type of the key is stored as
// the first char and the lookup subject is the rest of the key.
func (r *Resolver) lookupFunc(key string) func() (interface{}, error) {
	if len(key) == 0 {
		panic("lookupFunc with empty key")
	}

	var resolver DNSResolver = net.DefaultResolver
	if r.Resolver != nil {
		resolver = r.Resolver
	}

	switch key[0] {
	case 'h':
		return func() (interface{}, error) {
			ctx, cancel := r.getCtx()
			defer cancel()
			return resolver.LookupHost(ctx, key[1:])
		}
	case 'r':
		return func() (interface{}, error) {
			ctx, cancel := r.getCtx()
			defer cancel()
			return resolver.LookupAddr(ctx, key[1:])
		}
	default:
		panic("lookupFunc invalid key type: " + key)
	}
}

func (r *Resolver) getCtx() (ctx context.Context, cancel context.CancelFunc) {
	ctx = context.Background()
	if r.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
	} else {
		cancel = func() {}
	}
	return
}

func (r *Resolver) load(key string) (rrs []string, err error, found bool) {
	r.mu.RLock()
	var entry *cacheEntry
	entry, found = r.cache[key]
	if !found {
		r.mu.RUnlock()
		return
	}
	rrs = entry.rrs
	err = entry.err
	used := entry.used
	r.mu.RUnlock()
	if !used {
		r.mu.Lock()
		entry.used = true
		r.mu.Unlock()
	}
	return rrs, err, true
}

func (r *Resolver) storeLocked(key string, rrs []string, used bool, err error) {
	if entry, found := r.cache[key]; found {
		// Update existing entry in place
		entry.rrs = rrs
		entry.err = err
		entry.used = used
		return
	}
	r.cache[key] = &cacheEntry{
		rrs:  rrs,
		err:  err,
		used: used,
	}
}

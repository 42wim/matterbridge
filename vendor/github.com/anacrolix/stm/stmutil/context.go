package stmutil

import (
	"context"
	"sync"

	"github.com/anacrolix/stm"
)

var (
	mu      sync.Mutex
	ctxVars = map[context.Context]*stm.Var{}
)

// Returns an STM var that contains a bool equal to `ctx.Err != nil`, and a cancel function to be
// called when the user is no longer interested in the var.
func ContextDoneVar(ctx context.Context) (*stm.Var, func()) {
	mu.Lock()
	defer mu.Unlock()
	if v, ok := ctxVars[ctx]; ok {
		return v, func() {}
	}
	if ctx.Err() != nil {
		// TODO: What if we had read-only Vars? Then we could have a global one for this that we
		// just reuse.
		v := stm.NewBuiltinEqVar(true)
		return v, func() {}
	}
	v := stm.NewVar(false)
	go func() {
		<-ctx.Done()
		stm.AtomicSet(v, true)
		mu.Lock()
		delete(ctxVars, ctx)
		mu.Unlock()
	}()
	ctxVars[ctx] = v
	return v, func() {}
}

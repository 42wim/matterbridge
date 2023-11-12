package discv5

import (
	wenr "github.com/waku-org/go-waku/waku/v2/protocol/enr"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

// FilterPredicate is to create a Predicate using a custom function
func FilterPredicate(predicate func(*enode.Node) bool) Predicate {
	return func(iterator enode.Iterator) enode.Iterator {
		if predicate != nil {
			iterator = enode.Filter(iterator, predicate)
		}

		return iterator
	}
}

// FilterShard creates a Predicate that filters nodes that belong to a specific shard
func FilterShard(cluster, index uint16) Predicate {
	return func(iterator enode.Iterator) enode.Iterator {
		predicate := func(node *enode.Node) bool {
			rs, err := wenr.RelaySharding(node.Record())
			if err != nil || rs == nil {
				return false
			}
			return rs.Contains(cluster, index)
		}
		return enode.Filter(iterator, predicate)
	}
}

// FilterCapabilities creates a Predicate to filter nodes that support specific protocols
func FilterCapabilities(flags wenr.WakuEnrBitfield) Predicate {
	return func(iterator enode.Iterator) enode.Iterator {
		predicate := func(node *enode.Node) bool {
			enrField := new(wenr.WakuEnrBitfield)
			if err := node.Record().Load(enr.WithEntry(wenr.WakuENRField, &enrField)); err != nil {
				return false
			}

			if enrField == nil {
				return false
			}

			return *enrField&flags == flags
		}
		return enode.Filter(iterator, predicate)
	}
}

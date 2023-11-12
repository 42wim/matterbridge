package typeddata

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	bytes32Type, _ = abi.NewType("bytes32", "", nil)
	int256Type, _  = abi.NewType("int256", "", nil)

	errNotInteger = errors.New("not an integer")
)

// ValidateAndHash generates a hash of TypedData and verifies that chainId in the typed data matches currently selected chain.
func ValidateAndHash(typed TypedData, chain *big.Int) (common.Hash, error) {
	if err := typed.ValidateChainID(chain); err != nil {
		return common.Hash{}, err
	}

	return encodeData(typed)
}

// deps runs breadth-first traversal starting from target and collects all
// found composite dependencies types into result slice. target always will be first
// in the result array. all other dependencies are sorted alphabetically.
// for example: Z{c C, a A} A{c C} and the target is Z.
// result would be Z, A, B, C
func deps(target string, types Types) []string {
	unique := map[string]struct{}{}
	unique[target] = struct{}{}
	visited := []string{target}
	deps := []string{}
	for len(visited) > 0 {
		current := visited[0]
		fields := types[current]
		for i := range fields {
			f := fields[i]
			if _, defined := types[f.Type]; defined {
				if _, exist := unique[f.Type]; !exist {
					visited = append(visited, f.Type)
					unique[f.Type] = struct{}{}
				}
			}
		}
		visited = visited[1:]
		deps = append(deps, current)
	}
	sort.Slice(deps[1:], func(i, j int) bool {
		return deps[1:][i] < deps[1:][j]
	})
	return deps
}

func typeString(target string, types Types) string {
	b := new(bytes.Buffer)
	for _, dep := range deps(target, types) {
		b.WriteString(dep)
		b.WriteString("(")
		fields := types[dep]
		first := true
		for i := range fields {
			if !first {
				b.WriteString(",")
			} else {
				first = false
			}
			f := fields[i]
			b.WriteString(f.Type)
			b.WriteString(" ")
			b.WriteString(f.Name)
		}
		b.WriteString(")")
	}
	return b.String()
}

func typeHash(target string, types Types) (rst common.Hash) {
	return crypto.Keccak256Hash([]byte(typeString(target, types)))
}

func hashStruct(target string, data map[string]json.RawMessage, types Types) (rst common.Hash, err error) {
	fields := types[target]
	typeh := typeHash(target, types)
	args := abi.Arguments{{Type: bytes32Type}}
	vals := []interface{}{typeh}
	for i := range fields {
		f := fields[i]
		val, typ, err := toABITypeAndValue(f, data, types)
		if err != nil {
			return rst, err
		}
		vals = append(vals, val)
		args = append(args, abi.Argument{Name: f.Name, Type: typ})
	}
	packed, err := args.Pack(vals...)
	if err != nil {
		return rst, err
	}
	return crypto.Keccak256Hash(packed), nil
}

func toABITypeAndValue(f Field, data map[string]json.RawMessage, types Types) (val interface{}, typ abi.Type, err error) {
	if f.Type == "string" {
		var str string
		if err = json.Unmarshal(data[f.Name], &str); err != nil {
			return
		}
		return crypto.Keccak256Hash([]byte(str)), bytes32Type, nil
	} else if f.Type == "bytes" {
		var bytes hexutil.Bytes
		if err = json.Unmarshal(data[f.Name], &bytes); err != nil {
			return
		}
		return crypto.Keccak256Hash(bytes), bytes32Type, nil
	} else if _, exist := types[f.Type]; exist {
		var obj map[string]json.RawMessage
		if err = json.Unmarshal(data[f.Name], &obj); err != nil {
			return
		}
		val, err = hashStruct(f.Type, obj, types)
		if err != nil {
			return
		}
		return val, bytes32Type, nil
	}
	return atomicType(f, data)
}

func atomicType(f Field, data map[string]json.RawMessage) (val interface{}, typ abi.Type, err error) {
	typ, err = abi.NewType(f.Type, "", nil)
	if err != nil {
		return
	}
	if typ.T == abi.SliceTy || typ.T == abi.ArrayTy || typ.T == abi.FunctionTy {
		return val, typ, errors.New("arrays, slices and functions are not supported")
	} else if typ.T == abi.FixedBytesTy {
		return toFixedBytes(f, data[f.Name])
	} else if typ.T == abi.AddressTy {
		val, err = toAddress(f, data[f.Name])
	} else if typ.T == abi.IntTy || typ.T == abi.UintTy {
		return toInt(f, data[f.Name])
	} else if typ.T == abi.BoolTy {
		val, err = toBool(f, data[f.Name])
	} else {
		err = fmt.Errorf("type %s is not supported", f.Type)
	}
	return
}

func toFixedBytes(f Field, data json.RawMessage) (rst [32]byte, typ abi.Type, err error) {
	var bytes hexutil.Bytes
	if err = json.Unmarshal(data, &bytes); err != nil {
		return
	}
	typ = bytes32Type
	rst = [32]byte{}
	// reduce the length to the advertised size
	if len(bytes) > typ.Size {
		bytes = bytes[:typ.Size]
	}
	copy(rst[:], bytes)
	return rst, typ, nil
}

func toInt(f Field, data json.RawMessage) (val *big.Int, typ abi.Type, err error) {
	val = new(big.Int)
	if err = json.Unmarshal(data, &val); err != nil {
		var buf string
		err = json.Unmarshal(data, &buf)
		if err != nil {
			return
		}
		var ok bool
		val, ok = val.SetString(buf, 0)
		if !ok {
			err = errNotInteger
			return
		}
	}
	return val, int256Type, nil
}

func toAddress(f Field, data json.RawMessage) (rst common.Address, err error) {
	err = json.Unmarshal(data, &rst)
	return
}

func toBool(f Field, data json.RawMessage) (rst bool, err error) {
	err = json.Unmarshal(data, &rst)
	return
}

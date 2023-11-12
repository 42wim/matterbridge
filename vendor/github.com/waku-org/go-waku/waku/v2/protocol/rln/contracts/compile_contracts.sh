#!/bin/sh

# Generate solc_output.json
cat import.json | solcjs --bin --standard-json --base-path . > solc_output.json
tail -n +2 solc_output.json > solc_output.tmp  # Removes ">>> Cannot retry compilation with SMT because there are no SMT solvers available."
mv solc_output.tmp solc_output.json

# Generate abi and binary files for each smart contract
jq '.contracts."WakuRln.sol".WakuRln.abi' -r -c solc_output.json > WakuRln.abi
jq '.contracts."WakuRln.sol".WakuRln.evm.bytecode.object' -r -c solc_output.json > WakuRln.bin
jq '.contracts."WakuRlnRegistry.sol".WakuRlnRegistry.abi' -r -c solc_output.json > WakuRlnRegistry.abi
jq '.contracts."WakuRlnRegistry.sol".WakuRlnRegistry.evm.bytecode.object' -r -c solc_output.json > WakuRlnRegistry.bin
jq '.contracts."rln-contract/PoseidonHasher.sol".PoseidonHasher.abi' -r -c solc_output.json > PoseidonHasher.abi
jq '.contracts."rln-contract/PoseidonHasher.sol".PoseidonHasher.evm.bytecode.object' -r -c solc_output.json > PoseidonHasher.bin

# Generate golang types for each contract
abigen --abi ./WakuRln.abi --pkg contracts --type RLN --out ./rln.go --bin ./WakuRln.bin
abigen --abi ./WakuRlnRegistry.abi --pkg contracts --type RLNRegistry --out ./registry.go --bin ./WakuRlnRegistry.bin
abigen --abi ./PoseidonHasher.abi --pkg contracts --type PoseidonHasher --out ./poseidon.go --bin ./PoseidonHasher.bin

# Cleanup
rm *.bin
rm *.abi
rm solc_output.json
#!/usr/bin/env bash

set -euo pipefail

readonly src_dir=/tmp/reserve-stats-abi-gen/src/github.com/ethereum

go mod vendor

mkdir -p "$src_dir"
pushd "$src_dir"
[[ -d go-ethereum ]] || git clone https://github.com/favadi/go-ethereum.git
popd

cp -R "$src_dir"/go-ethereum/crypto/secp256k1/libsecp256k1 \
   ./vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/

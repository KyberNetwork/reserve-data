#!/usr/bin/env bash

set -euo pipefail

readonly src_dir=/tmp/reserve-stats-abi-gen/src/github.com/ethereum

go mod vendor

mkdir -p "$src_dir"
pushd "$src_dir"
[[ -d go-ethereum ]] || git clone https://github.com/ethereum/go-ethereum.git
popd

cp -R "$src_dir"/go-ethereum/crypto/secp256k1/libsecp256k1 \
   ./vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/

readonly src_usb=/tmp/reserve_data/vendor/usb
mkdir -p "$src_usb"
pushd "$src_usb"
[[ -d usb ]] || git clone https://github.com/karalabe/usb.git
popd

cp -R "$src_usb"/usb/hidapi ./vendor/github.com/karalabe/usb/hidapi
cp -R "$src_usb"/usb/libusb ./vendor/github.com/karalabe/usb/libusb
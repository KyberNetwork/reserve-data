#!/bin/bash

set -euo pipefail

cp /setting/* /go/src/github.com/KyberNetwork/reserve-data/cmd/
cp -R /migrations /go/src/github.com/KyberNetwork/reserve-data/cmd/
cd /go/src/github.com/KyberNetwork/reserve-data/cmd/
/cmd "$@"

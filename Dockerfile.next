# build stage
FROM golang:stretch AS build-env
COPY . /go/src/github.com/KyberNetwork/reserve-data
WORKDIR /go
RUN go install -v github.com/KyberNetwork/reserve-data/cmd

# final stage
FROM debian:stretch
ENV KYBER_EXCHANGES huobi,binance
COPY --from=build-env /go/bin/cmd /cmd
COPY ./entrypoint.sh /entrypoint.sh
COPY ./cmd/deposit_keystore ./cmd/keystore ./cmd/intermediate_account_keystore /setting/
COPY ./common/blockchain/ERC20.abi /go/src/github.com/KyberNetwork/reserve-data/common/blockchain/ERC20.abi
COPY ./blockchain/*.abi /go/src/github.com/KyberNetwork/reserve-data/blockchain/

RUN apt-get update && \
    apt install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/KyberNetwork/reserve-data/cmd
EXPOSE 8000

ENTRYPOINT ["/entrypoint.sh"]
CMD ["server", "--log-to-stdout"]

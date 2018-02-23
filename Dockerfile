# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/KyberNetwork/reserve-data

ENV KYBER_EXCHANGES huobi
WORKDIR /go/src/github.com/KyberNetwork/reserve-data
RUN go install -v github.com/KyberNetwork/reserve-data/cmd

EXPOSE 8000

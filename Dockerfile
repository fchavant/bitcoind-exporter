FROM golang AS builder
WORKDIR /go/src/github.com/hectorj/bitcoind-exporter
ADD . .
RUN git submodule update --init --recursive
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -linkmode external -extldflags -static" -v .

FROM alpine
WORKDIR /
COPY --from=builder /go/src/github.com/hectorj/bitcoind-exporter/bitcoind-exporter /
ENTRYPOINT ["/bitcoind-exporter"]
EXPOSE 8452
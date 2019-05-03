FROM golang:1.12 as builder

RUN apt-get install make git gcc libc-dev

WORKDIR /go/src/github.com/Factom-Asset-Tokens/fatd

# Use go mod in gopath
ENV GO111MODULE=on

# Populate all the source code
COPY . .

RUN make

# Now squash everything
# TODO: Use something lighter than golang?
FROM golang:1.12

COPY --from=builder /go/src/github.com/Factom-Asset-Tokens/fatd/fatd /go/bin/fatd
COPY --from=builder /go/src/github.com/Factom-Asset-Tokens/fatd/fat-cli /go/bin/fat-cli

ENTRYPOINT ["/go/bin/fatd"]
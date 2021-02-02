FROM golang:1.15.2-alpine3.12 as builder

ENV CGO_ENABLED=0

WORKDIR /go/src/couchdb-proxy
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GO111MODULE=on go build -i -v -a -installsuffix cgo -o app couchdb-proxy/cmd/server

###

FROM alpine:3.7
RUN apk add --no-cache ca-certificates=20190108-r0
WORKDIR /root
COPY --from=builder /go/src/couchdb-proxy/app .

CMD ["./app"]
FROM golang:1.11.1-alpine as build-env

ENV GO111MODULE on

RUN apk add --update build-base gcc wget git

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...
RUN go build -o bin/flowd cmd/flowd/main.go

FROM alpine
RUN apk add --no-cache ca-certificates

COPY --from=build-env /go/src/app/bin/flowd /usr/local/bin
CMD ["flowd"]

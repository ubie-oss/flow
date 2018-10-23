FROM golang:1.11.1 as build-env

ENV GO111MODULE on

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...
RUN go build -o bin/flowd cmd/flowd/main.go

FROM gcr.io/distroless/base
COPY --from=build-env /go/src/app/bin/flowd /
CMD ["/app"]

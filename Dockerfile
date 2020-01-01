FROM golang:1.13 as go
FROM gcr.io/distroless/base-debian10 as run

FROM go as build
WORKDIR /go/src/github.com/sakajunquality/flow

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o /go/bin/flowd ./cmd/flowd/main.go
RUN go build -o /go/bin/server ./cmd/server/main.go

FROM run
COPY --from=build /go/bin/flowd /usr/local/bin/flowd
COPY --from=build /go/bin/server /usr/local/bin/server
CMD ["server"]

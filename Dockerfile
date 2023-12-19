FROM golang:1.21 as go
FROM gcr.io/distroless/base-debian12 as run

FROM go as build
WORKDIR /go/src/github.com/sakajunquality/flow/v4

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/server 

FROM run
COPY --from=build /go/bin/server /usr/local/bin/server
CMD ["server"]

FROM golang:1.19.4-alpine as builder

WORKDIR /go/src/jq-api
COPY . .
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

RUN go mod vendor
RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o jq-api .

FROM stedolan/jq

COPY --from=builder /go/src/jq-api/jq-api .
ENTRYPOINT ["./jq-api"]

EXPOSE 8080
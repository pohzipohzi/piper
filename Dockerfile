FROM golang:1.14.0-alpine3.11

WORKDIR /go/src

COPY . .

RUN go build -o piper

ENTRYPOINT ["./piper"]

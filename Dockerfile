FROM golang:1.14.6-alpine

WORKDIR /go/src

COPY . .

RUN go install

ENTRYPOINT ["piper"]

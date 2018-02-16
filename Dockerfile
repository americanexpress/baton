FROM golang:1.9.3

RUN mkdir -p /go/src/baton
WORKDIR /go/src/baton

COPY . .

RUN go install

ENTRYPOINT ["baton"]

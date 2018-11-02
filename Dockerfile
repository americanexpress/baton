FROM golang:alpine as builder

# get deps ca-certs and git
RUN apk update && apk add git && apk add ca-certificates
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# create baton user
RUN adduser -D -g '' batonuser

# copy src and set working directory
COPY . $GOPATH/src/baton/
WORKDIR $GOPATH/src/baton/

# run dep 
RUN dep ensure --vendor-only

# disable support for c system libs, for use with scratch
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

# build and test our binary
RUN go test -v
RUN go build -a -installsuffix cgo -o /go/bin/baton

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/baton /go/bin/baton
ENTRYPOINT ["/go/bin/baton"]

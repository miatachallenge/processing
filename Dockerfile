FROM golang:1.10.3-stretch

COPY . /go/src/github.com/miatachallenge/processing
RUN go install -v github.com/miatachallenge/processing

FROM debian:stretch
COPY --from=0 /go/bin/processing /usr/bin/processing

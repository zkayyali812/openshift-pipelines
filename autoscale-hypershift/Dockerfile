FROM  golang:1.18 as builder

WORKDIR /go/src/github.com/stolostron/autoscale-hypershift
COPY . .

RUN make -f Makefile compile

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

COPY --from=builder /go/src/github.com/stolostron/autoscale-hypershift/action .

USER 1001

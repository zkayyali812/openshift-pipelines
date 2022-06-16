FROM  golang:1.18 as builder

WORKDIR /go/src/github.com/stolostron/kcp-ns-cleanup
COPY . .

RUN make -f Makefile compile

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

WORKDIR /src/

COPY --from=builder /go/src/github.com/stolostron/kcp-ns-cleanup/kcp-ns-cleanup .
COPY --from=builder /go/src/github.com/stolostron/kcp-ns-cleanup/allowlist.yaml .

RUN chmod -R 775 .

USER 1001

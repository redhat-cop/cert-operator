FROM alpine:3.6

RUN  apk update && \
     apk add libc6-compat && \
     apk add ca-certificates

RUN adduser -D cert-operator

USER cert-operator

ADD tmp/_output/bin/cert-operator /usr/local/bin/cert-operator

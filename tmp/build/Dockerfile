FROM alpine:3.6

RUN adduser -D cert-operator
USER cert-operator

ADD tmp/_output/bin/cert-operator /usr/local/bin/cert-operator

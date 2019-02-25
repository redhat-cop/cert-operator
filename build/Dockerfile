FROM alpine:3.6

RUN  apk update && \
     apk add libc6-compat && \
     apk add ca-certificates
#Uncomment below lines to copy venafi's public certificate to the docker container. 
ADD venafi-tpp-cert.pem /usr/local/share/ca-certificates/venafi-tpp-cert.pem
RUN update-ca-certificates

RUN adduser -D cert-operator

USER cert-operator

ADD tmp/_output/bin/cert-operator /usr/local/bin/cert-operator

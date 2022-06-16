FROM envoyproxy/envoy-alpine:v1.21.0

# Install Golang in alpine
#RUN apk update \
#&& apk add ca-certificates wget \ 
#&& update-ca-certificates
#RUN wget https://golang.org/dl/go1.17.5.linux-amd64.tar.gz \
#&& tar -C /usr/local -xzf go1.17.5.linux-amd64.tar.gz 
#ENV PATH=$PATH:/usr/local/go/bin

COPY envoy.yaml /etc/envoy/envoy.yaml
RUN chmod go+r /etc/envoy/envoy.yaml
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
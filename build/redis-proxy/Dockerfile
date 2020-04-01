FROM alpine:3.9

USER nobody

ADD build/_output/redis-proxy /usr/local/bin/redis-proxy

ENTRYPOINT ["redis-proxy"]

FROM alpine:3.9

USER nobody

ADD build/_output/redis-storage /usr/local/bin/redis-storage

ENTRYPOINT ["redis-storage"]

FROM alpine:3.9

USER nobody

ADD build/_output/redis-storage-node /usr/local/bin/redis-storage-node

ENTRYPOINT ["redis-storage-node"]

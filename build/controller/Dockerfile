FROM alpine:3.9

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/redis-storage-controller /usr/local/bin/redis-storage-controller

ENTRYPOINT ["redis-storage-controller"]

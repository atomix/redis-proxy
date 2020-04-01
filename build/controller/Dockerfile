FROM alpine:3.9

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/redis-controller /usr/local/bin/redis-controller

ENTRYPOINT ["redis-controller"]

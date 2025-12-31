# Use same Alpine version as minimalytics for layer sharing
FROM alpine:3.20

RUN apk add --no-cache haproxy

COPY haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg

EXPOSE 80

CMD ["haproxy", "-f", "/usr/local/etc/haproxy/haproxy.cfg"]

####################################################################################################
# base
####################################################################################################
FROM alpine:3.12.3 as base
RUN apk update && apk upgrade && \
    apk add ca-certificates && \
    apk --no-cache add tzdata

COPY dist/nats-source /bin/nats-source
RUN chmod +x /bin/nats-source

####################################################################################################
# nats-source
####################################################################################################
FROM scratch as nats-source
ARG ARCH
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=base /bin/nats-source /bin/nats-source
ENTRYPOINT [ "/bin/nats-source" ]

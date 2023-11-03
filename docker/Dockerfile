FROM registry.opensource.zalan.do/library/alpine-3.15:latest
LABEL maintainer="CPO Open-Source Project - CYBERTEC PostgreSQL International GmbH"

# We need root certificates to deal with teams api over https
RUN apk --no-cache add curl
RUN apk --no-cache add ca-certificates

COPY build/* /

RUN addgroup -g 1000 cpo
RUN adduser -D -u 1000 -G cpo -g 'Postgres Operator' pgo

USER 1000:1000

ENTRYPOINT ["/postgres-operator"]

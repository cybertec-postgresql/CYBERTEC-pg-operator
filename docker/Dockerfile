ARG BASE_IMAGE 
ARG VERSION=latest

FROM ${BASE_IMAGE} as builder
ARG PACKAGER
ARG VERSION

COPY  . /go/src/github.com/cybertec-postgresql/cybertec-pg-operator
WORKDIR /go/src/github.com/cybertec-postgresql/cybertec-pg-operator

ENV OPERATOR_LDFLAGS="-X=main.version=${VERSION}"

RUN ${PACKAGER} -y install --nodocs  https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm && ${PACKAGER} -y update && ${PACKAGER} -y install dpkg wget && ${PACKAGER} clean all;

RUN bash docker/build_operator.sh

FROM ${BASE_IMAGE}
ARG PACKAGER
LABEL maintainer="Opensource @ CYBERTEC <cpo-community@googlegroups.com>"

# We need root certificates to deal with teams api over https
RUN ${PACKAGER} -y update && ${PACKAGER} -y install ca-certificates && ${PACKAGER} clean all;

COPY --from=builder /go/src/github.com/cybertec-postgresql/cybertec-pg-operator/build/* /

RUN groupadd -g 1000 cpo
RUN useradd cpo -u 1000 -g 1000

USER 1000:1000

ENTRYPOINT ["/postgres-operator"]

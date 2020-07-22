FROM golang:1.14.5

ARG NOTARY_VERSION
ARG MIGRATE_VERSION
RUN test -n "$NOTARY_VERSION"
RUN test -n "$MIGRATE_VERSION"
ENV NOTARYPKG github.com/theupdateframework/notary
ENV MIGRATEPKG github.com/golang-migrate/migrate

RUN git clone -b $NOTARY_VERSION https://github.com/theupdateframework/notary.git /go/src/${NOTARYPKG}
WORKDIR /go/src/${NOTARYPKG}

RUN go install -tags pkcs11 \
    -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" ${NOTARYPKG}/cmd/notary-server 

RUN go install -tags pkcs11 \
    -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" ${NOTARYPKG}/cmd/notary-signer
RUN cp -r /go/src/${NOTARYPKG}/migrations/ / 

RUN git clone -b $MIGRATE_VERSION https://github.com/golang-migrate/migrate /go/src/${MIGRATEPKG}
WORKDIR /go/src/${MIGRATEPKG}

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN dep ensure -vendor-only

ENV DATABASES="postgres mysql redshift cassandra spanner cockroachdb clickhouse"
ENV SOURCES="file go_bindata github aws_s3 google_cloud_storage"

RUN go install -tags "$DATABASES $SOURCES" -ldflags="-X main.Version=${MIGRATE_VERSION}" ${MIGRATEPKG}/cli && mv /go/bin/cli /go/bin/migrate


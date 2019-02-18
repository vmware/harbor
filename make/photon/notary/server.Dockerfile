FROM photon:2.0
   
RUN tdnf install -y shadow sudo \
    && tdnf clean all \
    && groupadd -r -g 10000 notary \
    && useradd --no-log-init -r -g 10000 -u 10000 notary

COPY ./make/photon/notary/binary/notary-server /bin/notary-server
COPY ./make/photon/notary/server-start.sh /bin/server-start.sh
RUN chmod u+x /bin/notary-server /bin/server-start.sh
ENV SERVICE_NAME=notary_server
ENTRYPOINT [ "/bin/server-start.sh" ]

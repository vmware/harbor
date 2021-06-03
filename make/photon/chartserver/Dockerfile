ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-chartserver-base:${harbor_base_image_version}

COPY ./make/photon/chartserver/binary/chartm /home/chart/
COPY ./make/photon/chartserver/docker-entrypoint.sh /home/chart/
COPY ./make/photon/common/install_cert.sh /home/chart/

RUN chown -R chart:chart /etc/pki/tls/certs \
    && chown -R chart:chart /home/chart \
    && chmod u+x /home/chart/chartm \
    && chmod u+x /home/chart/docker-entrypoint.sh \
    && chmod u+x /home/chart/install_cert.sh

USER chart

WORKDIR /home/chart

ENTRYPOINT ["./docker-entrypoint.sh"]

VOLUME ["/chart_storage"]

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS http://localhost:9999/health || curl -k -sS https://localhost:9443/health || exit 1

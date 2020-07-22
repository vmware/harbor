ARG harbor_base_image_version
ARG harbor_base_namespace
FROM node:10.15.0 as nodeportal

WORKDIR /build_dir

ARG npm_registry=https://registry.npmjs.org

RUN apt-get update \
    && apt-get install -y --no-install-recommends python-yaml=3.12-1

COPY src/portal/package.json /build_dir
COPY src/portal/package-lock.json /build_dir
COPY src/portal/scripts /build_dir
COPY ./api/v2.0/legacy_swagger.yaml /build_dir/swagger.yaml
COPY ./api/v2.0/swagger.yaml /build_dir/swagger2.yaml
COPY ./api/swagger.yaml /build_dir/swagger3.yaml

RUN python -c 'import sys, yaml, json; y=yaml.load(sys.stdin.read()); print json.dumps(y)' < swagger.yaml > swagger.json
RUN python -c 'import sys, yaml, json; y=yaml.load(sys.stdin.read()); print json.dumps(y)' < swagger2.yaml > swagger2.json
RUN python -c 'import sys, yaml, json; y=yaml.load(sys.stdin.read()); print json.dumps(y)' < swagger3.yaml > swagger3.json

COPY ./LICENSE /build_dir
COPY src/portal /build_dir

ENV NPM_CONFIG_REGISTRY=${npm_registry}
RUN npm install --unsafe-perm
RUN npm run generate-build-timestamp
RUN node --max_old_space_size=2048 'node_modules/@angular/cli/bin/ng' build --prod

FROM ${harbor_base_namespace}/harbor-portal-base:${harbor_base_image_version}

COPY --from=nodeportal /build_dir/dist /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.yaml /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.json /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger2.json /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger3.json /usr/share/nginx/html
COPY --from=nodeportal /build_dir/LICENSE /usr/share/nginx/html

COPY make/photon/portal/nginx.conf /etc/nginx/nginx.conf

EXPOSE 8080
VOLUME /var/cache/nginx /var/log/nginx /run

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080 || exit 1
USER nginx
CMD ["nginx", "-g", "daemon off;"]


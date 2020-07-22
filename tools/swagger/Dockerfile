FROM golang:1.14.5

ARG SWAGGER_VERSION
RUN curl -fsSL -o /usr/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/$SWAGGER_VERSION/swagger_linux_amd64 && chmod +x /usr/bin/swagger

ENTRYPOINT ["/usr/bin/swagger"]
CMD ["--help"]

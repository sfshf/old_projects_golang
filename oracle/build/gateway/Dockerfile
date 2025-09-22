FROM golang:1.21 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/gateway-linux-amd64 ./cmd/gateway/main.go

FROM centos:8 AS release-stage
WORKDIR /app
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
# localizer configs
ADD configs/toml /etc/application/toml
# filebeat.yml
COPY build/gateway/filebeat.yml /etc/filebeat/filebeat.yml
# executable files
COPY --from=build-backend /app/dist/gateway-linux-amd64 /usr/local/bin/app-gateway
COPY build/gateway/entrypoint.sh /entrypoint.sh
RUN mkdir /log && \
    chmod +x /entrypoint.sh && \
    curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.12.0-x86_64.rpm && \
    rpm -vi filebeat-8.12.0-x86_64.rpm && rm filebeat-8.12.0-x86_64.rpm

ENTRYPOINT ["/entrypoint.sh"]
CMD ["app-gateway"]

# http port
ARG GATEWAY_HTTP_PORT
ARG GATEWAY_TLS_PORT
ARG GATEWAY_GRPC_PORT

EXPOSE ${GATEWAY_HTTP_PORT}
EXPOSE ${GATEWAY_TLS_PORT}
EXPOSE ${GATEWAY_GRPC_PORT}

STOPSIGNAL SIGTERM
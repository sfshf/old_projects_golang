FROM golang:1.23 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/grpc-linux-amd64 ./cmd/grpc/main.go

FROM centos:8 AS release-stage
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
ENV AWS_ACCESS_KEY_ID=
ENV AWS_SECRET_ACCESS_KEY=
ENV AWS_REGION=us-east-1
ADD configs/toml /etc/application/toml
COPY build/filebeat.yml /etc/filebeat/filebeat.yml
COPY build/entrypoint.sh /entrypoint.sh
RUN mkdir /log && \
    chmod +x /entrypoint.sh && \
    curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.12.0-x86_64.rpm && \
    rpm -vi filebeat-8.12.0-x86_64.rpm && rm filebeat-8.12.0-x86_64.rpm

# executable files
COPY --from=build-backend /app/dist/grpc-linux-amd64 /usr/local/bin/app-grpc

ENTRYPOINT ["/entrypoint.sh"]
CMD ["app-grpc"]

ARG GRPC_SERVER_PORT
ARG GRPC_METRIC_PORT
ARG MYSQL_METRIC_PORT

# grpc server ports
## 6000 -- grpc port
## 6001 -- grpc prom port
## 6002 -- grpc mysql prom port
EXPOSE $GRPC_SERVER_PORT
EXPOSE $GRPC_METRIC_PORT
EXPOSE $MYSQL_METRIC_PORT

STOPSIGNAL SIGTERM
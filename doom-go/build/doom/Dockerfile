FROM golang:1.23 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags "-extldflags '-Wl,--allow-multiple-definition'" -o dist/grpc-linux-amd64 ./cmd/doom/main.go

FROM centos:8 AS release-stage
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
COPY build/doom/filebeat.yml /etc/filebeat/filebeat.yml
ADD configs/errorlang /etc/application/errorlang
ADD configs/cryptocurrency /etc/application/cryptocurrency
COPY build/doom/entrypoint.sh /entrypoint.sh
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

EXPOSE ${GRPC_SERVER_PORT}
EXPOSE ${GRPC_METRIC_PORT}
EXPOSE ${MYSQL_METRIC_PORT}

STOPSIGNAL SIGTERM
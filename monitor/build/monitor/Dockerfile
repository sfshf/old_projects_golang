FROM golang:1.22 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags "-extldflags '-Wl,--allow-multiple-definition'" -o dist/app-linux-amd64 ./cmd/monitor/main.go

FROM centos:latest AS release-stage
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
COPY build/monitor/filebeat.yml /etc/filebeat/filebeat.yml
ADD configs/errorlang /etc/application/errorlang
COPY build/monitor/entrypoint.sh /entrypoint.sh
RUN mkdir /log && \
    chmod +x /entrypoint.sh && \
    curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.12.0-x86_64.rpm && \
    rpm -vi filebeat-8.12.0-x86_64.rpm && rm filebeat-8.12.0-x86_64.rpm

# executable files
COPY --from=build-backend /app/dist/app-linux-amd64 /usr/local/bin/app

ENTRYPOINT ["/entrypoint.sh"]
CMD ["app"]

STOPSIGNAL SIGTERM
FROM golang:1.23 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/grpc-linux-amd64 ./cmd/grpc/main.go

FROM node:18-alpine AS build-ui
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
ARG env
COPY ./ui /app/
RUN apk add --no-cache libc6-compat
RUN if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
    elif [ -f package-lock.json ]; then npm ci; \
    elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i; \
    else echo "Lockfile not found." && exit 1; \
    fi
RUN cp .env.$env .env.local && \
    yarn build

FROM centos:8 AS release-stage
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
COPY build/grpc/filebeat.yml /etc/filebeat/filebeat.yml
ADD configs/toml /etc/application/toml
COPY build/grpc/entrypoint.sh /entrypoint.sh
RUN mkdir /log && \
    chmod +x /entrypoint.sh && \
    curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.12.0-x86_64.rpm && \
    rpm -vi filebeat-8.12.0-x86_64.rpm && rm filebeat-8.12.0-x86_64.rpm
# executable files
COPY --from=build-backend /app/dist/grpc-linux-amd64 /usr/local/bin/app-grpc
# static files
COPY --from=build-ui /app/dist /var/www/dist
ENTRYPOINT ["/entrypoint.sh"]
CMD ["app-grpc"]
ARG GRPC_SERVER_PORT
ARG GRPC_METRIC_PORT
ARG MYSQL_METRIC_PORT
ARG HTTP_SERVER_PORT

# grpc server ports
## 9100 -- grpc port
## 9101 -- grpc prom port
## 9102 -- grpc mysql prom port
EXPOSE $GRPC_SERVER_PORT
EXPOSE $GRPC_METRIC_PORT
EXPOSE $MYSQL_METRIC_PORT
# http server ports
## 9110 -- http port
EXPOSE $HTTP_SERVER_PORT

STOPSIGNAL SIGTERM
FROM golang:1.21 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/console-linux-amd64 ./cmd/console/main.go

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
# RUN cp .env.$env .env.local && yarn build
RUN yarn build

FROM centos:8 AS release-stage
WORKDIR /app
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
# localizer configs
ADD configs/toml /etc/application/toml
# required proto files
COPY ./proto/ /app/proto/
# filebeat.yml
COPY build/console/filebeat.yml /etc/filebeat/filebeat.yml
# executable files
COPY --from=build-backend /app/dist/console-linux-amd64 /usr/local/bin/app-console
COPY build/console/entrypoint.sh /entrypoint.sh
RUN mkdir /log && \
    chmod +x /entrypoint.sh && \
    curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.12.0-x86_64.rpm && \
    rpm -vi filebeat-8.12.0-x86_64.rpm && rm filebeat-8.12.0-x86_64.rpm
# static files
COPY --from=build-ui /app/dist /var/www/dist
ENTRYPOINT ["/entrypoint.sh"]
CMD ["app-console"]

# http port
ARG CONSOLE_HTTP_PORT
# grpc port
ARG CONSOLE_GRPC_PORT

EXPOSE ${CONSOLE_HTTP_PORT}
EXPOSE ${CONSOLE_GRPC_PORT}

STOPSIGNAL SIGTERM
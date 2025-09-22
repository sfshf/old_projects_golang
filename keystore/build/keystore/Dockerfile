FROM golang:1.21 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/http-linux-amd64 ./cmd/keystore/main.go

FROM centos:8 AS release-stage
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
# executable files
COPY --from=build-backend /app/dist/http-linux-amd64 /usr/local/bin/app-http
COPY --from=build-backend /app/build/keystore/.env /etc/keystore/.env
CMD ["app-http"]

# grpc server ports
EXPOSE 1111

STOPSIGNAL SIGTERM
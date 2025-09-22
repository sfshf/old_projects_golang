FROM golang:1.23 AS build-backend
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/grpc-linux-amd64 ./cmd/notification/main.go

FROM centos:8 AS release-stage
LABEL maintainer="Luo <luoxianmingg@gmail.com>"
ADD configs/toml /etc/application/toml
# executable files
COPY --from=build-backend /app/dist/grpc-linux-amd64 /usr/local/bin/app-grpc
CMD ["app-grpc"]
# grpc server ports
EXPOSE 7010
EXPOSE 7011
EXPOSE 7012

STOPSIGNAL SIGTERM
FROM golang:1.19 AS build-stage

WORKDIR /app
COPY . /app/
# RUN git config --global url.https://github@github.com.insteadof https://github.com
RUN GOSUMDB=off go mod tidy
RUN CGO_ENABLED=0  GOARCH="amd64" GOOS=linux go build -o /app/app

# Run the tests in the container
# FROM build-stage AS run-test-stage
# RUN go test -v ./...

# Deploy the application binary into a lean image
FROM centos:latest AS release-stage

WORKDIR /app

COPY --from=build-stage /app/app /app/app
COPY conf-release.json /app/conf.json
COPY level.toml /app/level.toml
RUN mkdir -p /app/download/csv
RUN mkdir -p /app/download/bundle
RUN mkdir -p /app/backup
EXPOSE 9001

ENTRYPOINT ["/app/app"]
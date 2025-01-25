FROM golang:1.23.5 AS build-stage

WORKDIR /app

COPY ./cache_node/go.mod ./cache_node/go.sum ./

RUN go mod download

COPY ./cache_node/*.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ./cache_node

# Run the tests in the container
# FROM build-stage AS run-test-stage

# RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage ./app/cache_node ./cache_node

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/cache_node"]

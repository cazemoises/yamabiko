FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /core-api ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=build /core-api ./core-api
COPY --from=build /app/migrations ./migrations
EXPOSE 8080
CMD ["./core-api"]

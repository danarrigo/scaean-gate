FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/sync-worker ./cmd/worker

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /bin/sync-worker /usr/local/bin/sync-worker
USER app
ENTRYPOINT ["sync-worker"]

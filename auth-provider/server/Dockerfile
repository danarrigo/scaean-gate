FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/auth-provider ./cmd/api

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /bin/auth-provider /usr/local/bin/auth-provider
USER app
EXPOSE 8080
ENTRYPOINT ["auth-provider"]

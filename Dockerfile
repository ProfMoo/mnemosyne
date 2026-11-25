FROM golang:1.26.11-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go run ./cmd/frontend-export \
    && CGO_ENABLED=0 go build -trimpath -o /mnemosyne ./cmd/mnemosyne

FROM alpine:3.24
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN addgroup -S mnemosyne \
    && adduser -S -G mnemosyne mnemosyne \
    && mkdir /data \
    && chown mnemosyne:mnemosyne /data
COPY --from=build /mnemosyne /usr/local/bin/mnemosyne
COPY --from=build /src/bin/web /app/web
WORKDIR /data
USER mnemosyne
VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["mnemosyne", "-config", "config/docker.yaml"]

FROM node:26.13.1-alpine AS frontend
RUN npm install --global npm@12.2.3
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
COPY api/ /src/api/
COPY cmd/mnemosyne/config/config.yaml /src/cmd/mnemosyne/config/config.yaml
RUN npm run generate && npm run check

FROM golang:1.26.11-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN sed -i 's|^frontendDir:.*|frontendDir: "/app/frontend/dist"|' cmd/mnemosyne/config/config.yaml \
    && CGO_ENABLED=0 go build -trimpath -o /mnemosyne ./cmd/mnemosyne

FROM alpine:3.24
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN addgroup -S mnemosyne \
    && adduser -S -G mnemosyne mnemosyne \
    && mkdir /data \
    && chown mnemosyne:mnemosyne /data
COPY --from=backend /mnemosyne /usr/local/bin/mnemosyne
COPY --from=frontend /src/frontend/dist /app/frontend/dist
WORKDIR /data
USER mnemosyne
VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["mnemosyne"]

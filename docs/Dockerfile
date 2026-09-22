FROM golang:1.27-alpine AS wasm-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOARCH=wasm GOOS=js go build -o web/app.wasm .

FROM golang:1.27-alpine AS static-generator
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=wasm-builder /app/web/ ./web
RUN CGO_ENABLED=0 go build -o local-server .
RUN ./local-server static -o dist
RUN cp -r web/ dist/
RUN mv dist/web/data dist/data

FROM ghcr.io/paulopiriquito/hog-runtime:v2.0.0

COPY --from=static-generator /app/dist /srv/web
COPY --chown=hog:hog server /etc/hog/

# The gateway (server/gateway.yaml) listens on port 80.
# hog-runtime runs as non-root uid 10001, so binding a port below 1024
# needs the unprivileged-port sysctl lowered. Run this image with:
#   podman run --sysctl net.ipv4.ip_unprivileged_port_start=0 -p 8080:80 <image>
EXPOSE 80

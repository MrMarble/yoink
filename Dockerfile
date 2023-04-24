FROM golang:1.20-alpine

# copy over the binary from the first stage
COPY yoink /app/yoink

WORKDIR "/app"
ENTRYPOINT ["/app/yoink", "-c /config.yaml"]
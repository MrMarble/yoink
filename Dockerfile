FROM golang:1.20-alpine
ENTRYPOINT ["/yoink", "-c /config.yaml"]
COPY yoink /
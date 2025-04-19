FROM scratch

# copy over the binary from the first stage
COPY yoink /app/yoink

WORKDIR "/app"
ENTRYPOINT ["/app/yoink", "-c /config.yaml"]
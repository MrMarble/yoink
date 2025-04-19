FROM gcr.io/distroless/static-debian12

# copy over the binary from the first stage
COPY yoink /yoink

CMD ["/yoink", "-c", "/config.yaml"]

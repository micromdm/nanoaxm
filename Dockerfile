FROM gcr.io/distroless/static

ARG TARGETOS TARGETARCH

COPY nanoaxm-$TARGETOS-$TARGETARCH /app/nanoaxm

EXPOSE 9005

VOLUME ["/app/db"]

WORKDIR /app

ENTRYPOINT ["/app/nanoaxm"]

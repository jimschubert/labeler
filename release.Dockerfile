FROM gcr.io/distroless/base-debian12
ARG APP_NAME
COPY /${APP_NAME} /app
ENTRYPOINT ["/app"]

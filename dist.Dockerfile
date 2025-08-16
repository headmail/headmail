# Runtime image that consumes pre-built binaries placed into dist/<platform>/headmail by the CI job.
# TARGETPLATFORM is provided by buildx (e.g., "linux/amd64" or "linux/arm64").
FROM alpine:3.22

ARG TARGETPLATFORM

RUN mkdir -p /app/ /app/data /app/config

# Copy the binary built by CI into the image. Buildx will set TARGETPLATFORM (contains a slash).
# The CI places binaries at dist/<os>/<arch>/headmail (e.g. dist/linux/amd64/headmail).
COPY dist/${TARGETPLATFORM}/headmail /app/headmail

RUN chmod +x /app/headmail && \
    addgroup -S headmail && adduser -S -G headmail headmail && \
    chown headmail:headmail -R /app/data

ENV HEADMAIL_DATABASE_TYPE="sqlite"
ENV HEADMAIL_DATABASE_URL="file:/app/data/data.db?cache=shared&mode=rwc"

USER headmail
WORKDIR /app

EXPOSE 8080
EXPOSE 8081

VOLUME /app/data

CMD ["/app/headmail"]

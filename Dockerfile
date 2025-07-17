# Herald Docker Image
FROM alpine:latest

# Install dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Create herald user
RUN adduser -D -s /bin/sh herald

# Set up working directory
WORKDIR /app

# Copy Herald binary (provided by GoReleaser)
COPY herald /usr/local/bin/herald

# Copy configuration template
COPY .heraldrc /app/.heraldrc.template

# Make herald executable
RUN chmod +x /usr/local/bin/herald

# Switch to herald user
USER herald

# Set up git safe directory for mounted volumes
RUN git config --global --add safe.directory /app

# Default command
ENTRYPOINT ["herald"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="Herald"
LABEL org.opencontainers.image.description="Release management tool using conventional commits"
LABEL org.opencontainers.image.url="https://github.com/herald"
LABEL org.opencontainers.image.documentation="https://github.com/herald/blob/main/README.md"
LABEL org.opencontainers.image.source="https://github.com/herald"
LABEL org.opencontainers.image.licenses="MIT" 
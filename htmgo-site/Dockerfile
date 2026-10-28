# Build from the repository root so the relative replace directives in go.mod
# resolve inside the build context, and the image is built from the same
# commit as the framework:
#
#   docker build -f htmgo-site/Dockerfile .
#
# htmgo-site/Dockerfile.dockerignore limits the context to what this image needs.

# Stage 1: Install the npm packages Tailwind CSS plugins are loaded from
FROM node:22-bookworm-slim AS node-deps

WORKDIR /src/htmgo-site
COPY htmgo-site/package.json htmgo-site/package-lock.json ./
RUN npm ci

# Stage 2: Build the Go binary
FROM golang:1.27 AS builder

# The app is not part of go.work; resolve modules from its own go.mod.
ENV GOWORK=off
# Static binary, independent of the builder's glibc version.
ENV CGO_ENABLED=0

# Mirror the repository layout so the relative replace paths resolve.
WORKDIR /src

# Build the htmgo CLI from the same commit as the app.
COPY cli/htmgo/ cli/htmgo/
RUN cd cli/htmgo && go install .

# Copy go.mod and go.sum files first so module downloads are cached
COPY framework/go.mod framework/go.sum framework/
COPY tools/html-to-htmgo/go.mod tools/html-to-htmgo/go.sum tools/html-to-htmgo/
COPY htmgo-site/go.mod htmgo-site/go.sum htmgo-site/
RUN cd htmgo-site && go mod download

# Copy the source code into the container
COPY framework/ framework/
COPY tools/html-to-htmgo/ tools/html-to-htmgo/
COPY htmgo-site/ htmgo-site/
COPY --from=node-deps /src/htmgo-site/node_modules htmgo-site/node_modules

WORKDIR /src/htmgo-site

# Build the Go binary for Linux
RUN LOG_LEVEL=debug htmgo build


# Stage 3: Create the smallest possible image
FROM gcr.io/distroless/base-debian12

# Set the working directory inside the container
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /src/htmgo-site/dist .

# Expose the necessary port (replace with your server port)
EXPOSE 3000

# Command to run the binary
CMD ["./htmgo-site"]

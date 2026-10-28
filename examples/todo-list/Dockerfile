# Build from the repository root so the relative replace directives in go.mod
# resolve inside the build context, and the image is built from the same
# commit as the framework:
#
#   docker build -f examples/todo-list/Dockerfile .
#
# examples/todo-list/Dockerfile.dockerignore limits the context to what this image needs.

# Stage 1: Build the Go binary
FROM golang:1.27 AS builder

# The app is not part of go.work; resolve modules from its own go.mod.
ENV GOWORK=off

# Mirror the repository layout so the relative replace paths resolve.
WORKDIR /src

# Build the htmgo CLI from the same commit as the app.
COPY cli/htmgo/ cli/htmgo/
RUN cd cli/htmgo && go install .

# Copy go.mod and go.sum files first so module downloads are cached
COPY framework/go.mod framework/go.sum framework/
COPY examples/todo-list/go.mod examples/todo-list/go.sum examples/todo-list/
RUN cd examples/todo-list && go mod download

# Copy the source code into the container
COPY framework/ framework/
COPY examples/todo-list/ examples/todo-list/

WORKDIR /src/examples/todo-list

# Build the Go binary for Linux
RUN CGO_ENABLED=0 LOG_LEVEL=debug htmgo build

# Relink with cgo statically for go-sqlite3.
RUN CGO_ENABLED=1 GOOS=linux go build -o ./dist -a -ldflags '-linkmode external -extldflags "-static"' .


# Stage 2: Create the smallest possible image
FROM gcr.io/distroless/base-debian12

# Set the working directory inside the container
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /src/examples/todo-list/dist .

# Expose the necessary port (replace with your server port)
EXPOSE 3000

# Command to run the binary
CMD ["./todolist"]

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
# # Avoid dynamic linking of libc, since we are using a different deployment image
# # that might have a different version of libc.
# ENV CGO_ENABLED=0

# Install make
RUN apk add --no-cache make git

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# # Copy Makefile separately to verify it exists
# COPY Makefile ./
# RUN ls -la && cat Makefile

# Copy the rest of the source code
COPY . .

# Build using make
RUN make build-linux

# Runtime stage
FROM alpine:3.18

# # Add basic tools for CLI access
# RUN apk add --no-cache curl tzdata bash

# ENV GOTRACEBACK=single

WORKDIR /app
# Copy the binary and resources
COPY --from=builder /app/myapp /app/myapp
COPY --from=builder /app/internal/resources /app/internal/resources
COPY --from=builder /app/view/templates /app/view/templates
# Make sure files are writable for task updates
RUN chmod 777 /app/internal/resources
# List files to debug
RUN ls -la

RUN ls -la internal/resources
# Environment variables
ENV PORT=8080 \
    TASKS_FILE=/app/internal/resources/tasks.json \
    USERS_FILE=/app/internal/resources/users.json

EXPOSE 8080

CMD ["/app/myapp", "-web", "/app/internal/resources/tasks.json"]
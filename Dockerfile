# Build stage: using Go to build the application
FROM golang:alpine3.20 AS build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image
COPY go.mod ./
COPY go.sum ./

# download Go modules and dependencies
RUN go mod download

# copy the entire directory containing Go files
COPY . .

# compile application
RUN go build -o /executor

# Final stage: using Playwright image and installing Go and Playwright Go client
FROM node:20-bookworm

# Set the Go version to install
ENV GO_VERSION=1.23.0
# Download and install Go manually (latest version)
RUN apt-get update && \
    apt-get install -y wget tar && \
    wget https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz && \
    rm go$GO_VERSION.linux-amd64.tar.gz

# Add Go to PATH
ENV PATH=$PATH:/usr/local/go/bin

# Install the Go Playwright client
RUN go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps

# Copy the built application from the build stage
COPY --from=build /executor /executor

# Set the working directory and copy additional files
WORKDIR /

# Expose the application's port
EXPOSE 8081

# Command to run the application
CMD [ "/executor" ]

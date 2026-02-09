#!/bin/bash

# Build the auth-service binary
echo "Building auth-service..."
go build -o auth-service cmd/main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Run with: ./auth-service"
else
    echo "Build failed!"
    exit 1
fi

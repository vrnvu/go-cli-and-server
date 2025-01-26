#!/bin/bash

# Set the base URL based on the environment
case "$ENV" in
  dev)  BASE_URL="http://localhost:8080" ;;
  pre)  BASE_URL="http://pre.example.com" ;;
  pro)  BASE_URL="http://pro.example.com" ;;
  *)    echo "Invalid ENV value. Must be 'dev', 'pre', or 'pro'." && exit 1 ;;
esac

echo "Using BASE_URL: $BASE_URL"

# Helper functions
check_status() { [ "$1" -eq "$2" ] || { echo "Expected status $2 but got $1"; exit 1; } }
check_json() { [ "$1" = "$2" ] || { echo "Expected JSON $2 but got $1"; exit 1; } }

echo "Testing health endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
check_status $response 200
echo "Health check passed."

echo "Integration tests passed."

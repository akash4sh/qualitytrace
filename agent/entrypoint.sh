#!/bin/sh

# Validate that TRACETEST_API_KEY is not empty
if [ -z "$TRACETEST_API_KEY" ]; then
  echo "Error: TRACETEST_API_KEY environment variable is empty." >&2
  exit 1
fi

# Execute qualitytrace with the API key and any additional arguments
exec qualitytrace start --api-key "$TRACETEST_API_KEY" "$@"

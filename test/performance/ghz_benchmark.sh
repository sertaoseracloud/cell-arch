#!/bin/bash
# ghz benchmark script for sidecar performance
set -e

PORT=50051
BINARY=./sidecar
HEALTH_URL="http://localhost:$PORT/healthz"

# 1. Build sidecar
echo "Building sidecar..."
go build -o "$BINARY" ./cmd/sidecar

# 2. Start sidecar in background
echo "Starting sidecar on port $PORT..."
"$BINARY" &
SIDECAR_PID=$!
cleanup() {
  echo "Stopping sidecar (PID $SIDECAR_PID)..."
  kill $SIDECAR_PID 2>/dev/null || true
  wait $SIDECAR_PID 2>/dev/null || true
}
trap cleanup EXIT

# 3. Wait for health endpoint
echo "Waiting for sidecar to become ready..."
for i in $(seq 1 30); do
  if curl -s -o /dev/null -w "%{http_code}" "$HEALTH_URL" | grep -q 200; then
    echo "Sidecar is ready."
    break
  fi
  sleep 1
done

# 4. Run ghz benchmark
echo "Running ghz benchmark..."
GHZ_OUTPUT=$(ghz --insecure \
  -n 100000 \
  -c 100 \
  --latency \
  -m GET \
  -u cell-arch.GreeterService/Greet \
  "localhost:$PORT" 2>&1)

echo "$GHZ_OUTPUT"

# 5. Parse latency and throughput
P95=$(echo "$GHZ_OUTPUT" | grep -o '"p95":[0-9.]*' | cut -d: -f2)
RPS=$(echo "$GHZ_OUTPUT" | grep -o '"rps":[0-9.]*' | cut -d: -f2)

echo "p95 latency: ${P95} ms"
echo "Throughput: ${RPS} rps"

# 6. Validate thresholds (convert to ms where needed)
if [ -z "$P95" ] || [ -z "$RPS" ]; then
  echo "ERROR: Could not parse ghz output."
  exit 1
fi

# Fail if p95 > 5 ms or rps < 1000
if [ "$(echo "$P95 > 5" | bc)" -eq 1 ]; then
  echo "FAIL: p95 latency ${P95} ms exceeds 5 ms threshold."
  exit 1
fi

if [ "$(echo "$RPS < 1000" | bc)" -eq 1 ]; then
  echo "FAIL: Throughput ${RPS} rps below 1000 rps threshold."
  exit 1
fi

echo "Benchmark PASSED."
exit 0

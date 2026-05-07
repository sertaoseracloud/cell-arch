#!/usr/bin/env bash
# scripts/validate_phase2.sh
# Automates Phase2 validation: protobuf generation, build, test, and coverage check.
set -euo pipefail

echo "=== Phase2 Validation Script ==="

# 1. Check Go toolchain
if ! command -v go &> /dev/null; then
    echo "ERROR: Go toolchain not found. Please install Go."
    exit 1
fi
echo "Go version: $(go version)"

# 2. Check protoc
if ! command -v protoc &> /dev/null; then
    echo "ERROR: protoc not found. Please install Protocol Buffers compiler."
    exit 1
fi
echo "protoc version: $(protoc --version)"

# 3. Regenerate protobuf stubs
echo ""
echo "--- Regenerating protobuf stubs ---"
protoc --go_out=. --go-grpc_out=. proto/task.proto
echo "Protobuf stubs regenerated."

# 4. Build sidecar
echo ""
echo "--- Building sidecar ---"
go build ./cmd/sidecar
echo "Build SUCCESS."

# 5. Run tests with coverage
echo ""
echo "--- Running tests with coverage ---"
go test -cover ./... 2>&1 | tee /tmp/phase2_test_output.txt
echo ""
echo "Coverage summary:"
grep -E "coverage:|no test files" /tmp/phase2_test_output.txt || true

# 6. Check coverage thresholds (>=80%)
echo ""
echo "--- Checking coverage thresholds (>=80%) ---"
FAILED=0
while IFS= read -r line; do
    if [[ "$line" =~ ^(ok|FAIL).*coverage: ]]; then
        pkg=$(echo "$line" | awk '{print $2}')
        cov=$(echo "$line" | grep -oP 'coverage: \K[0-9.]+')
        if (( $(echo "$cov < 80" | bc -l) )); then
            echo "WARNING: $pkg coverage $cov% is below 80%"
            FAILED=1
        else
            echo "PASS: $pkg coverage $cov%"
        fi
    fi
done < /tmp/phase2_test_output.txt

if [ $FAILED -eq 1 ]; then
    echo "Some packages have coverage below 80%."
else
    echo "All packages meet the 80% coverage threshold."
fi

# 7. Summary
echo ""
echo "=== Phase2 Validation Complete ==="
echo "Next step: Run /gsd-validate-phase 2 in OpenClaude to update VALIDATION.md"

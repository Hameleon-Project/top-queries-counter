set -euo pipefail

URL="${1:-http://localhost:8080/api/top?n=10}"

echo "== Go benchmarks (in-memory store) =="
go test ./internal/store -bench=. -benchmem -count=3

if command -v hey >/dev/null 2>&1; then
  echo ""
  echo "== HTTP load: hey -z 10s -c 50 $URL =="
  hey -z 10s -c 50 "$URL"
else
  echo ""
  echo "Install hey for HTTP load test: go install github.com/rakyll/hey@latest"
fi

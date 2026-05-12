#!/usr/bin/env bash
set -euo pipefail

echo "==> start benchmark"

echo "==> starting containers"
docker compose up -d

echo "==> waiting for app"
until curl -fsS http://localhost:8080/_info > /dev/null; do
  sleep 1
done

echo "==> seeding data"
go run bench/seed.go

extract_json_field() {
  local json="$1"
  local field="$2"
  echo "$json" | sed -n "s/.*\"$field\":\"\([^\"]*\)\".*/\1/p"
}

echo "==> getting admin token"
TOKEN_RESPONSE=$(curl -sS -X POST http://localhost:8080/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}')

TOKEN=$(extract_json_field "$TOKEN_RESPONSE" "token")

if [ -z "${TOKEN}" ]; then
  echo "failed to get token"
  echo "response: $TOKEN_RESPONSE"
  exit 1
fi

echo "==> getting first room id"
ROOMS_RESPONSE=$(curl -sS -H "Authorization: Bearer ${TOKEN}" \
  http://localhost:8080/rooms/list)

ROOM_ID=$(extract_json_field "$ROOMS_RESPONSE" "id")

if [ -z "${ROOM_ID}" ]; then
  echo "failed to get room id"
  echo "response: $ROOMS_RESPONSE"
  exit 1
fi

DATE=$(date -u -d tomorrow +%Y-%m-%d 2>/dev/null || date -u -v+1d +%Y-%m-%d)

echo "==> benchmark target"
echo "GET /rooms/${ROOM_ID}/slots/list?date=${DATE}"

echo "==> running hey"
hey -n 5000 -c 50 \
  -H "Authorization: Bearer ${TOKEN}" \
  "http://localhost:8080/rooms/${ROOM_ID}/slots/list?date=${DATE}" \
  > bench_results.txt 2>&1

echo "==> benchmark completed"
cat bench_results.txt
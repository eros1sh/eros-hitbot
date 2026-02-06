#!/bin/bash
# Test script for ErosHit Distributed Mode (Bash)
# Bu script tek makinede test etmek için kullanılır

set -e

MASTER_PORT=8080
WORKER_COUNT=2
TASK_COUNT=10
MASTER_URL="http://localhost:$MASTER_PORT"
SECRET_KEY="test-secret-key"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║     ErosHit Distributed Mode - Local Test                  ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"

cd "$ROOT_DIR"

# Build binaries
echo -e "${GREEN}Building binaries...${NC}"
mkdir -p bin
go build -o bin/master cmd/eroshit/master.go
go build -o bin/worker cmd/eroshit/worker.go
echo -e "${GREEN}Binaries built successfully${NC}"
echo ""

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    pkill -f "bin/master" 2>/dev/null || true
    pkill -f "bin/worker" 2>/dev/null || true
}

trap cleanup EXIT

# Start Master
echo -e "${GREEN}Starting Master on port $MASTER_PORT...${NC}"
./bin/master -bind "127.0.0.1:$MASTER_PORT" -secret "$SECRET_KEY" &
MASTER_PID=$!
sleep 2

# Check if master is running
if ! curl -s "$MASTER_URL/api/v1/master/status" > /dev/null; then
    echo -e "${RED}Master failed to start!${NC}"
    exit 1
fi

echo -e "${GREEN}Master is running (PID: $MASTER_PID)${NC}"
echo ""

# Start Workers
echo -e "${GREEN}Starting $WORKER_COUNT Workers...${NC}"
WORKER_PIDS=()
for i in $(seq 1 $WORKER_COUNT); do
    ./bin/worker -master "$MASTER_URL" -secret "$SECRET_KEY" -concurrency 5 &
    WORKER_PIDS+=($!)
    echo "Worker $i started (PID: ${WORKER_PIDS[-1]})"
done

sleep 2
echo ""

# Submit test tasks
echo -e "${GREEN}Submitting $TASK_COUNT test tasks...${NC}"
for i in $(seq 1 $TASK_COUNT); do
    curl -s -X POST "$MASTER_URL/api/v1/master/task/submit" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $SECRET_KEY" \
        -d "{\"url\": \"https://httpbin.org/get\", \"session_id\": \"test-session-$i\"}" > /dev/null
    echo "  Task $i submitted"
done

echo ""
echo -e "${CYAN}Monitoring progress...${NC}"
echo ""

# Monitor progress
for i in $(seq 1 30); do
    sleep 2
    
    STATS=$(curl -s "$MASTER_URL/api/v1/master/stats" -H "Authorization: Bearer $SECRET_KEY")
    TOTAL=$(echo "$STATS" | grep -o '"total_tasks":[0-9]*' | cut -d: -f2)
    COMPLETED=$(echo "$STATS" | grep -o '"completed_tasks":[0-9]*' | cut -d: -f2)
    FAILED=$(echo "$STATS" | grep -o '"failed_tasks":[0-9]*' | cut -d: -f2)
    PENDING=$(echo "$STATS" | grep -o '"pending_tasks":[0-9]*' | cut -d: -f2)
    WORKERS=$(echo "$STATS" | grep -o '"active_workers":[0-9]*' | cut -d: -f2)
    
    printf "\rWorkers: %s | Total: %s | Completed: %s | Failed: %s | Pending: %s" \
        "$WORKERS" "$TOTAL" "$COMPLETED" "$FAILED" "$PENDING"
    
    if [ "$PENDING" = "0" ] && [ "$((COMPLETED + FAILED))" = "$TOTAL" ] && [ "$TOTAL" != "0" ]; then
        echo ""
        break
    fi
done

echo ""

# Get workers list
echo -e "${CYAN}Connected Workers:${NC}"
curl -s "$MASTER_URL/api/v1/master/workers" -H "Authorization: Bearer $SECRET_KEY" | \
    grep -o '"id":"[^"]*","hostname":"[^"]*","total_tasks":[0-9]*' | \
    while read line; do
        ID=$(echo "$line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        HOST=$(echo "$line" | grep -o '"hostname":"[^"]*"' | cut -d'"' -f4)
        TASKS=$(echo "$line" | grep -o '"total_tasks":[0-9]*' | cut -d: -f2)
        echo "  - $ID ($HOST): $TASKS tasks"
    done

echo ""
echo -e "${GREEN}Test completed!${NC}"
echo ""

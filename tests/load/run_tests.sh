#!/bin/bash

# k6 Load Test Runner for Gorax Platform
# Runs load tests and collects results

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
WS_URL="${WS_URL:-ws://localhost:8080}"
TEST_USER_EMAIL="${TEST_USER_EMAIL:-loadtest@example.com}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-loadtest123}"
SCENARIO="${SCENARIO:-load}"
OUTPUT_DIR="${OUTPUT_DIR:-./results}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}Error: k6 is not installed${NC}"
    echo "Install k6: https://k6.io/docs/get-started/installation/"
    exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Banner
echo -e "${BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║         Gorax Platform Load Testing Suite                ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Function to run a test
run_test() {
    local test_file=$1
    local test_name=$2
    local result_file="${OUTPUT_DIR}/${test_name}_${TIMESTAMP}.json"
    local summary_file="${OUTPUT_DIR}/${test_name}_${TIMESTAMP}_summary.txt"

    echo -e "${YELLOW}Running ${test_name}...${NC}"
    echo "  Base URL: ${BASE_URL}"
    echo "  Scenario: ${SCENARIO}"
    echo "  Output: ${result_file}"
    echo ""

    # Export environment variables for k6
    export BASE_URL
    export WS_URL
    export TEST_USER_EMAIL
    export TEST_USER_PASSWORD
    export SCENARIO
    export TEST_ENV="${TEST_ENV:-local}"

    # Run k6 test
    if k6 run \
        --out json="${result_file}" \
        --summary-export="${summary_file}" \
        "${test_file}"; then
        echo -e "${GREEN}✓ ${test_name} completed successfully${NC}\n"
        return 0
    else
        echo -e "${RED}✗ ${test_name} failed${NC}\n"
        return 1
    fi
}

# Function to display help
show_help() {
    cat << EOF
Usage: $0 [OPTIONS] [TEST]

Run k6 load tests for the Gorax platform.

OPTIONS:
    -h, --help              Show this help message
    -u, --url URL           Base URL (default: http://localhost:8080)
    -w, --ws-url URL        WebSocket URL (default: ws://localhost:8080)
    -s, --scenario NAME     Test scenario: smoke|load|stress|spike|soak (default: load)
    -o, --output DIR        Output directory (default: ./results)
    -a, --all               Run all tests
    --email EMAIL           Test user email
    --password PASSWORD     Test user password

TESTS:
    workflow                Test workflow CRUD operations
    execution               Test workflow execution
    webhook                 Test webhook ingestion
    websocket               Test WebSocket connections
    auth                    Test authentication endpoints
    all                     Run all tests (default)

SCENARIOS:
    smoke                   Minimal load (1 VU, 1 minute)
    load                    Normal load (10 VUs ramped)
    stress                  High load (up to 100 VUs)
    spike                   Sudden traffic spike
    soak                    Extended duration (1 hour)

EXAMPLES:
    # Run all tests with load scenario
    $0 --scenario load

    # Run only workflow tests with stress scenario
    $0 workflow --scenario stress

    # Run smoke test on staging environment
    $0 --url https://staging.gorax.io --scenario smoke

    # Run soak test with custom output directory
    $0 --scenario soak --output /tmp/load-results

    # Run all tests
    $0 all

EOF
}

# Parse command line arguments
TEST_TO_RUN="all"
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -w|--ws-url)
            WS_URL="$2"
            shift 2
            ;;
        -s|--scenario)
            SCENARIO="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --email)
            TEST_USER_EMAIL="$2"
            shift 2
            ;;
        --password)
            TEST_USER_PASSWORD="$2"
            shift 2
            ;;
        -a|--all)
            TEST_TO_RUN="all"
            shift
            ;;
        workflow|execution|webhook|websocket|auth|all)
            TEST_TO_RUN="$1"
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Run tests based on selection
echo "Configuration:"
echo "  Base URL: ${BASE_URL}"
echo "  WebSocket URL: ${WS_URL}"
echo "  Scenario: ${SCENARIO}"
echo "  Test: ${TEST_TO_RUN}"
echo ""

if [[ "$TEST_TO_RUN" == "all" ]] || [[ "$TEST_TO_RUN" == "workflow" ]]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test "workflow_api.js" "workflow_api"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
fi

if [[ "$TEST_TO_RUN" == "all" ]] || [[ "$TEST_TO_RUN" == "execution" ]]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test "execution_api.js" "execution_api"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
fi

if [[ "$TEST_TO_RUN" == "all" ]] || [[ "$TEST_TO_RUN" == "webhook" ]]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test "webhook_trigger.js" "webhook_trigger"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
fi

if [[ "$TEST_TO_RUN" == "all" ]] || [[ "$TEST_TO_RUN" == "websocket" ]]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test "websocket_connections.js" "websocket_connections"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
fi

if [[ "$TEST_TO_RUN" == "all" ]] || [[ "$TEST_TO_RUN" == "auth" ]]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if run_test "auth_endpoints.js" "auth_endpoints"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
fi

# Summary
echo -e "${BLUE}"
echo "═══════════════════════════════════════════════════════════"
echo "                    Test Summary"
echo "═══════════════════════════════════════════════════════════"
echo -e "${NC}"
echo "Total Tests:  ${TOTAL_TESTS}"
echo -e "Passed:       ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed:       ${RED}${FAILED_TESTS}${NC}"
echo ""
echo "Results saved to: ${OUTPUT_DIR}"
echo ""

# Generate combined report if all tests were run
if [[ "$TEST_TO_RUN" == "all" ]]; then
    echo "Generating combined report..."
    REPORT_FILE="${OUTPUT_DIR}/combined_report_${TIMESTAMP}.html"

    cat > "$REPORT_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gorax Load Test Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        h1, h2, h3 { color: #2c3e50; }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 8px;
            margin-bottom: 30px;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        .status-pass { color: #27ae60; }
        .status-fail { color: #e74c3c; }
        table {
            width: 100%;
            border-collapse: collapse;
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ecf0f1;
        }
        th {
            background: #667eea;
            color: white;
            font-weight: 600;
        }
        tr:hover { background: #f8f9fa; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Gorax Platform Load Test Report</h1>
        <p>Generated: $(date)</p>
        <p>Scenario: ${SCENARIO}</p>
    </div>

    <div class="summary">
        <div class="card">
            <h3>Total Tests</h3>
            <div class="metric">${TOTAL_TESTS}</div>
        </div>
        <div class="card">
            <h3>Passed</h3>
            <div class="metric status-pass">${PASSED_TESTS}</div>
        </div>
        <div class="card">
            <h3>Failed</h3>
            <div class="metric status-fail">${FAILED_TESTS}</div>
        </div>
    </div>

    <div class="card">
        <h2>Test Results</h2>
        <p>Detailed results are available in the JSON files in the results directory:</p>
        <ul>
            <li>workflow_api_${TIMESTAMP}.json</li>
            <li>execution_api_${TIMESTAMP}.json</li>
            <li>webhook_trigger_${TIMESTAMP}.json</li>
            <li>websocket_connections_${TIMESTAMP}.json</li>
            <li>auth_endpoints_${TIMESTAMP}.json</li>
        </ul>
    </div>
</body>
</html>
EOF

    echo -e "${GREEN}Combined report generated: ${REPORT_FILE}${NC}"
fi

# Exit with appropriate code
if [[ $FAILED_TESTS -gt 0 ]]; then
    exit 1
else
    exit 0
fi

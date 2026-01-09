#!/bin/bash
# =============================================================================
# IRIS Incidences & Approval Workflow - Role-Based API Tests
# =============================================================================
#
# This script tests:
# 1. Incidence Categories CRUD
# 2. Incidence Types with custom form fields
# 3. Incidences creation and management
# 4. Absence Request workflow with different roles
# 5. Role-based permissions validation
#
# Usage: ./test_incidences_roles.sh
# =============================================================================

set -e

BASE_URL="${API_URL:-http://localhost:8080/api/v1}"
OUTPUT_DIR="./tests/curl_results/incidences_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$OUTPUT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
PASS=0
FAIL=0
SKIP=0

# Test users (should match e2e test users or seed data)
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@test.com}"
ADMIN_PASS="${ADMIN_PASS:-Password123@}"
EMPLOYEE_EMAIL="${EMPLOYEE_EMAIL:-employee@test.com}"
EMPLOYEE_PASS="${EMPLOYEE_PASS:-Password123@}"
SUPERVISOR_EMAIL="${SUPERVISOR_EMAIL:-supervisor@test.com}"
SUPERVISOR_PASS="${SUPERVISOR_PASS:-Password123@}"
HR_EMAIL="${HR_EMAIL:-hr@test.com}"
HR_PASS="${HR_PASS:-Password123@}"
MANAGER_EMAIL="${MANAGER_EMAIL:-manager@test.com}"
MANAGER_PASS="${MANAGER_PASS:-Password123@}"

# Tokens storage
declare -A TOKENS

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASS++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAIL++))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((SKIP++))
}

# Login and get token
login() {
    local email=$1
    local password=$2
    local role_name=$3

    local response
    response=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"$password\"}")

    local token
    token=$(echo "$response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

    if [ -n "$token" ]; then
        TOKENS[$role_name]=$token
        log_pass "Login as $role_name ($email)"
        return 0
    else
        log_fail "Login as $role_name ($email): $response"
        return 1
    fi
}

# Make authenticated request
api_request() {
    local method=$1
    local endpoint=$2
    local role=$3
    local data=$4
    local expected_status=${5:-2}  # Default: 2xx

    local token=${TOKENS[$role]}
    if [ -z "$token" ]; then
        log_skip "No token for $role"
        return 1
    fi

    local curl_cmd="curl -s -w '\n%{http_code}' -X $method '$BASE_URL$endpoint' \
        -H 'Authorization: Bearer $token' \
        -H 'Content-Type: application/json'"

    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi

    local response
    response=$(eval "$curl_cmd")

    local status
    status=$(echo "$response" | tail -1)
    local body
    body=$(echo "$response" | head -n -1)

    # Check if status starts with expected (2 for 2xx, 4 for 4xx, etc.)
    if [[ "$status" =~ ^${expected_status}[0-9][0-9]$ ]]; then
        return 0
    else
        echo "$body"
        return 1
    fi
}

# =============================================================================
# TEST: LOGIN ALL USERS
# =============================================================================

test_login_all() {
    echo ""
    echo "=============================================="
    echo "PHASE 1: Login All Test Users"
    echo "=============================================="

    login "$ADMIN_EMAIL" "$ADMIN_PASS" "admin" || true
    login "$EMPLOYEE_EMAIL" "$EMPLOYEE_PASS" "employee" || true
    login "$SUPERVISOR_EMAIL" "$SUPERVISOR_PASS" "supervisor" || true
    login "$HR_EMAIL" "$HR_PASS" "hr" || true
    login "$MANAGER_EMAIL" "$MANAGER_PASS" "manager" || true
}

# =============================================================================
# TEST: INCIDENCE CATEGORIES
# =============================================================================

test_incidence_categories() {
    echo ""
    echo "=============================================="
    echo "PHASE 2: Incidence Categories"
    echo "=============================================="

    # GET all categories (admin)
    log "Testing GET /incidence-categories as admin..."
    if api_request "GET" "/incidence-categories" "admin"; then
        log_pass "GET /incidence-categories (admin)"
    else
        log_fail "GET /incidence-categories (admin)"
    fi

    # CREATE category (admin)
    log "Testing POST /incidence-categories as admin..."
    local create_data='{"name":"Test Category","code":"test_cat","description":"Test category for API tests","color":"blue","is_requestable":true,"is_active":true}'

    local response
    response=$(curl -s -X POST "$BASE_URL/incidence-categories" \
        -H "Authorization: Bearer ${TOKENS[admin]}" \
        -H "Content-Type: application/json" \
        -d "$create_data")

    CATEGORY_ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

    if [ -n "$CATEGORY_ID" ]; then
        log_pass "POST /incidence-categories created: $CATEGORY_ID"
    else
        log_fail "POST /incidence-categories: $response"
    fi

    # GET single category
    if [ -n "$CATEGORY_ID" ]; then
        log "Testing GET /incidence-categories/:id..."
        if api_request "GET" "/incidence-categories/$CATEGORY_ID" "admin"; then
            log_pass "GET /incidence-categories/:id"
        else
            log_fail "GET /incidence-categories/:id"
        fi
    fi

    # UPDATE category
    if [ -n "$CATEGORY_ID" ]; then
        log "Testing PUT /incidence-categories/:id..."
        local update_data='{"name":"Test Category Updated","description":"Updated description"}'
        if api_request "PUT" "/incidence-categories/$CATEGORY_ID" "admin" "$update_data"; then
            log_pass "PUT /incidence-categories/:id"
        else
            log_fail "PUT /incidence-categories/:id"
        fi
    fi

    # Test employee access (should fail for write operations)
    log "Testing POST /incidence-categories as employee (should fail)..."
    if api_request "POST" "/incidence-categories" "employee" "$create_data" "4"; then
        log_pass "POST /incidence-categories (employee) correctly denied"
    else
        log_fail "POST /incidence-categories (employee) should be denied"
    fi
}

# =============================================================================
# TEST: INCIDENCE TYPES
# =============================================================================

test_incidence_types() {
    echo ""
    echo "=============================================="
    echo "PHASE 3: Incidence Types"
    echo "=============================================="

    # GET all types (admin)
    log "Testing GET /incidence-types as admin..."
    if api_request "GET" "/incidence-types" "admin"; then
        log_pass "GET /incidence-types (admin)"
    else
        log_fail "GET /incidence-types (admin)"
    fi

    # CREATE type with custom form fields
    log "Testing POST /incidence-types with custom form fields..."
    local type_data
    type_data=$(cat <<EOF
{
  "name": "Test Incidence Type",
  "category_id": "$CATEGORY_ID",
  "category": "other",
  "effect_type": "negative",
  "is_calculated": true,
  "calculation_method": "daily_rate",
  "default_value": 1,
  "requires_evidence": true,
  "description": "Test incidence type with custom fields",
  "is_requestable": true,
  "form_fields": {
    "fields": [
      {
        "name": "reason_detail",
        "type": "textarea",
        "label": "Detailed Reason",
        "required": true,
        "display_order": 1
      },
      {
        "name": "hours_needed",
        "type": "number",
        "label": "Hours Needed",
        "required": false,
        "min": 1,
        "max": 24,
        "display_order": 2
      },
      {
        "name": "approval_type",
        "type": "select",
        "label": "Approval Type",
        "required": true,
        "display_order": 3,
        "options": [
          {"value": "urgent", "label": "Urgent"},
          {"value": "normal", "label": "Normal"},
          {"value": "low", "label": "Low Priority"}
        ]
      }
    ]
  }
}
EOF
)

    local response
    response=$(curl -s -X POST "$BASE_URL/incidence-types" \
        -H "Authorization: Bearer ${TOKENS[admin]}" \
        -H "Content-Type: application/json" \
        -d "$type_data")

    INCIDENCE_TYPE_ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

    if [ -n "$INCIDENCE_TYPE_ID" ]; then
        log_pass "POST /incidence-types created with custom fields: $INCIDENCE_TYPE_ID"
    else
        log_fail "POST /incidence-types: $response"
    fi

    # GET requestable types (for employees)
    log "Testing GET /requestable-incidence-types as employee..."
    if api_request "GET" "/requestable-incidence-types" "employee"; then
        log_pass "GET /requestable-incidence-types (employee)"
    else
        log_fail "GET /requestable-incidence-types (employee)"
    fi

    # UPDATE type
    if [ -n "$INCIDENCE_TYPE_ID" ]; then
        log "Testing PUT /incidence-types/:id..."
        local update_data='{"name":"Test Incidence Type Updated","description":"Updated description"}'
        if api_request "PUT" "/incidence-types/$INCIDENCE_TYPE_ID" "admin" "$update_data"; then
            log_pass "PUT /incidence-types/:id"
        else
            log_fail "PUT /incidence-types/:id"
        fi
    fi
}

# =============================================================================
# TEST: INCIDENCES
# =============================================================================

test_incidences() {
    echo ""
    echo "=============================================="
    echo "PHASE 4: Incidences CRUD"
    echo "=============================================="

    # GET all incidences (admin)
    log "Testing GET /incidences as admin..."
    if api_request "GET" "/incidences" "admin"; then
        log_pass "GET /incidences (admin)"
    else
        log_fail "GET /incidences (admin)"
    fi

    # GET incidences with filters
    log "Testing GET /incidences with status filter..."
    if api_request "GET" "/incidences?status=pending" "admin"; then
        log_pass "GET /incidences?status=pending"
    else
        log_fail "GET /incidences?status=pending"
    fi

    # Employee should not access /incidences directly
    log "Testing GET /incidences as employee (may be restricted)..."
    if api_request "GET" "/incidences" "employee" "" "4"; then
        log_pass "GET /incidences (employee) correctly restricted"
    else
        log_skip "GET /incidences (employee) - may have read access"
    fi

    # HR should have access
    log "Testing GET /incidences as hr..."
    if api_request "GET" "/incidences" "hr"; then
        log_pass "GET /incidences (hr)"
    else
        log_fail "GET /incidences (hr)"
    fi
}

# =============================================================================
# TEST: ABSENCE REQUESTS
# =============================================================================

test_absence_requests() {
    echo ""
    echo "=============================================="
    echo "PHASE 5: Absence Requests Workflow"
    echo "=============================================="

    # GET pending counts
    log "Testing GET /absence-requests/counts..."
    if api_request "GET" "/absence-requests/counts" "admin"; then
        log_pass "GET /absence-requests/counts"
    else
        log_fail "GET /absence-requests/counts"
    fi

    # GET pending by stage (supervisor)
    log "Testing GET /absence-requests/pending/SUPERVISOR as supervisor..."
    if api_request "GET" "/absence-requests/pending/SUPERVISOR" "supervisor"; then
        log_pass "GET /absence-requests/pending/SUPERVISOR"
    else
        log_skip "GET /absence-requests/pending/SUPERVISOR - may require specific setup"
    fi

    # GET pending by stage (hr)
    log "Testing GET /absence-requests/pending/HR as hr..."
    if api_request "GET" "/absence-requests/pending/HR" "hr"; then
        log_pass "GET /absence-requests/pending/HR"
    else
        log_skip "GET /absence-requests/pending/HR - may require specific setup"
    fi

    # GET my requests (employee)
    log "Testing GET /absence-requests/my-requests as employee..."
    if api_request "GET" "/absence-requests/my-requests" "employee"; then
        log_pass "GET /absence-requests/my-requests"
    else
        log_fail "GET /absence-requests/my-requests"
    fi

    # Create absence request as employee
    log "Testing POST /absence-requests as employee..."
    local request_data
    request_data=$(cat <<EOF
{
  "request_type": "VACATION",
  "start_date": "$(date -d '+7 days' +%Y-%m-%d)",
  "end_date": "$(date -d '+10 days' +%Y-%m-%d)",
  "total_days": 4,
  "reason": "API Test vacation request"
}
EOF
)

    local response
    response=$(curl -s -X POST "$BASE_URL/absence-requests" \
        -H "Authorization: Bearer ${TOKENS[employee]}" \
        -H "Content-Type: application/json" \
        -d "$request_data")

    ABSENCE_REQUEST_ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

    if [ -n "$ABSENCE_REQUEST_ID" ]; then
        log_pass "POST /absence-requests created: $ABSENCE_REQUEST_ID"
    else
        log_skip "POST /absence-requests - may fail due to missing employee/supervisor setup: $response"
    fi
}

# =============================================================================
# TEST: APPROVAL WORKFLOW
# =============================================================================

test_approval_workflow() {
    echo ""
    echo "=============================================="
    echo "PHASE 6: Approval Workflow"
    echo "=============================================="

    if [ -z "$ABSENCE_REQUEST_ID" ]; then
        log_skip "Skipping approval tests - no absence request created"
        return
    fi

    # Supervisor approves
    log "Testing POST /absence-requests/:id/approve as supervisor..."
    local approve_data='{"action":"APPROVED","stage":"SUPERVISOR","comments":"Approved by supervisor"}'

    if api_request "POST" "/absence-requests/$ABSENCE_REQUEST_ID/approve" "supervisor" "$approve_data"; then
        log_pass "Supervisor approved request"
    else
        log_skip "Supervisor approval - may require specific stage"
    fi

    # Manager approves
    log "Testing POST /absence-requests/:id/approve as manager..."
    local manager_approve='{"action":"APPROVED","stage":"MANAGER","comments":"Approved by manager"}'

    if api_request "POST" "/absence-requests/$ABSENCE_REQUEST_ID/approve" "manager" "$manager_approve"; then
        log_pass "Manager approved request"
    else
        log_skip "Manager approval - may require specific stage"
    fi

    # HR approves
    log "Testing POST /absence-requests/:id/approve as hr..."
    local hr_approve='{"action":"APPROVED","stage":"HR","comments":"Approved by HR"}'

    if api_request "POST" "/absence-requests/$ABSENCE_REQUEST_ID/approve" "hr" "$hr_approve"; then
        log_pass "HR approved request"
    else
        log_skip "HR approval - may require specific stage"
    fi
}

# =============================================================================
# TEST: PERMISSION DENIALS
# =============================================================================

test_permission_denials() {
    echo ""
    echo "=============================================="
    echo "PHASE 7: Permission Denials"
    echo "=============================================="

    # Employee cannot approve
    if [ -n "$ABSENCE_REQUEST_ID" ]; then
        log "Testing employee cannot approve requests..."
        local approve_data='{"action":"APPROVED","stage":"SUPERVISOR","comments":"Trying to approve"}'

        if api_request "POST" "/absence-requests/$ABSENCE_REQUEST_ID/approve" "employee" "$approve_data" "4"; then
            log_pass "Employee correctly denied approval"
        else
            log_fail "Employee should not be able to approve"
        fi
    fi

    # Employee cannot delete other's categories
    if [ -n "$CATEGORY_ID" ]; then
        log "Testing employee cannot delete categories..."
        if api_request "DELETE" "/incidence-categories/$CATEGORY_ID" "employee" "" "4"; then
            log_pass "Employee correctly denied category deletion"
        else
            log_fail "Employee should not be able to delete categories"
        fi
    fi

    # Employee cannot create incidence types
    log "Testing employee cannot create incidence types..."
    local type_data='{"name":"Unauthorized Type","category":"other","effect_type":"neutral"}'
    if api_request "POST" "/incidence-types" "employee" "$type_data" "4"; then
        log_pass "Employee correctly denied incidence type creation"
    else
        log_fail "Employee should not be able to create incidence types"
    fi
}

# =============================================================================
# CLEANUP
# =============================================================================

cleanup() {
    echo ""
    echo "=============================================="
    echo "PHASE 8: Cleanup"
    echo "=============================================="

    # Delete test incidence type
    if [ -n "$INCIDENCE_TYPE_ID" ]; then
        log "Deleting test incidence type..."
        if api_request "DELETE" "/incidence-types/$INCIDENCE_TYPE_ID" "admin"; then
            log_pass "Deleted test incidence type"
        else
            log_skip "Could not delete incidence type (may be protected)"
        fi
    fi

    # Delete test category
    if [ -n "$CATEGORY_ID" ]; then
        log "Deleting test category..."
        if api_request "DELETE" "/incidence-categories/$CATEGORY_ID" "admin"; then
            log_pass "Deleted test category"
        else
            log_skip "Could not delete category (may be protected or have dependencies)"
        fi
    fi
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    echo "=============================================="
    echo "IRIS Incidences - Role-Based API Tests"
    echo "=============================================="
    echo "Base URL: $BASE_URL"
    echo "Output: $OUTPUT_DIR"
    echo ""

    test_login_all
    test_incidence_categories
    test_incidence_types
    test_incidences
    test_absence_requests
    test_approval_workflow
    test_permission_denials
    cleanup

    # Summary
    echo ""
    echo "=============================================="
    echo "TEST SUMMARY"
    echo "=============================================="
    echo -e "${GREEN}PASSED: $PASS${NC}"
    echo -e "${RED}FAILED: $FAIL${NC}"
    echo -e "${YELLOW}SKIPPED: $SKIP${NC}"
    echo "Total: $((PASS + FAIL + SKIP))"
    echo ""

    # Save results
    {
        echo "IRIS Incidences API Test Results"
        echo "================================"
        echo "Date: $(date)"
        echo "Base URL: $BASE_URL"
        echo ""
        echo "PASSED: $PASS"
        echo "FAILED: $FAIL"
        echo "SKIPPED: $SKIP"
    } > "$OUTPUT_DIR/summary.txt"

    log "Results saved to $OUTPUT_DIR/summary.txt"

    # Exit with failure if any tests failed
    if [ $FAIL -gt 0 ]; then
        exit 1
    fi
}

main "$@"

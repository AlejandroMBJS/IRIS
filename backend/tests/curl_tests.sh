#!/bin/bash

# IRIS Payroll Backend API Tests with cURL
# =========================================

set -e # Exit immediately if a command exits with a non-zero status

BASE_URL="http://localhost:8080/api/v1"
TEST_DIR="./tests/curl_results"
mkdir -p "$TEST_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variables to store tokens and IDs
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USER_ID=""

# Counter for test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print section headers
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

# Function to print test result
print_result() {
    local test_name="$1"
    local status="$2"
    local message="$3" # Optional message for failed tests
    
    if [ "$status" -eq 0 ]; then
        echo -e "${GREEN}✓ $test_name${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ $test_name ($message)${NC}"
        ((TESTS_FAILED++))
    fi
}

# Function to make API request and save response
make_request() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local auth_header="$5"
    
    # Replace spaces in test_name for filenames
    local sanitized_test_name=$(echo "$test_name" | tr ' ' '_')
    local response_file="$TEST_DIR/${sanitized_test_name}_response.json"
    local status_file="$TEST_DIR/${sanitized_test_name}_status.txt"
    
    echo "Test: $test_name" > "$status_file"
    echo "Endpoint: $endpoint" >> "$status_file"
    echo "Method: $method" >> "$status_file"
    
    # Build curl command
    CMD="curl -s -w '\n%{http_code}' -X "$method" '$BASE_URL$endpoint'"
    
    if [ -n "$data" ]; then
        CMD="$CMD -H 'Content-Type: application/json' -d '$data'"
    fi
    
    if [ -n "$auth_header" ]; then
        CMD="$CMD -H 'Authorization: Bearer $auth_header'"
    fi
    
    # Execute and capture response
    echo "Command: $CMD" >> "$status_file"
    RESPONSE=$(eval "$CMD")
    
    # Split response and status code
    STATUS_CODE=$(echo "$RESPONSE" | tail -n1)
    RESP_BODY=$(echo "$RESPONSE" | sed '$d')
    
    echo "$RESP_BODY" > "$response_file"
    echo "Status Code: $STATUS_CODE" >> "$status_file"
    
    # Return status code
    return "$STATUS_CODE"
}

# --- Main Test Execution ---

print_header "Authentication Tests"

# Test 1: User Registration
USER_EMAIL="test_user_$(date +%s%N)@example.com"
REG_DATA='{"company_name": "TestCorp", "company_rfc": "TCS123456ABC", "email": "'$USER_EMAIL'", "password": "Password123!", "role": "admin", "full_name": "Test Admin"}'
make_request "Register Admin User" "POST" "/auth/register" "$REG_DATA" ""
REG_HTTP_STATUS=$?
if [ "$REG_HTTP_STATUS" -eq 201 ]; then
    ACCESS_TOKEN=$(jq -r '.access_token' "$TEST_DIR/Register_Admin_User_response.json")
    REFRESH_TOKEN=$(jq -r '.refresh_token' "$TEST_DIR/Register_Admin_User_response.json")
    USER_ID=$(jq -r '.user.id' "$TEST_DIR/Register_Admin_User_response.json")
    echo "$ACCESS_TOKEN" > "$TEST_DIR/access_token.txt"
    echo "$REFRESH_TOKEN" > "$TEST_DIR/refresh_token.txt"
    print_result "Register Admin User" 0
else
    REG_MESSAGE=$(jq -r '.message' "$TEST_DIR/Register_Admin_User_response.json" || echo "Unknown error")
    print_result "Register Admin User" 1 "HTTP $REG_HTTP_STATUS: $REG_MESSAGE"
    echo -e "${RED}Registration failed, cannot proceed with other auth tests. Exiting.${NC}"
    exit 1 # Exit script on critical failure
fi

# Test 2: User Login
LOGIN_DATA='{"email": "'$USER_EMAIL'", "password": "Password123!"}'
make_request "Login User" "POST" "/auth/login" "$LOGIN_DATA" ""
LOGIN_HTTP_STATUS=$?
if [ "$LOGIN_HTTP_STATUS" -eq 200 ]; then
    ACCESS_TOKEN=$(jq -r '.access_token' "$TEST_DIR/Login_User_response.json")
    REFRESH_TOKEN=$(jq -r '.refresh_token' "$TEST_DIR/Login_User_response.json")
    echo "$ACCESS_TOKEN" > "$TEST_DIR/access_token.txt"
    echo "$REFRESH_TOKEN" > "$TEST_DIR/refresh_token.txt"
    print_result "Login User" 0
else
    LOGIN_MESSAGE=$(jq -r '.message' "$TEST_DIR/Login_User_response.json" || echo "Unknown error")
    print_result "Login User" 1 "HTTP $LOGIN_HTTP_STATUS: $LOGIN_MESSAGE"
fi

# Test 3: Get User Profile (Protected)
if [ -n "$ACCESS_TOKEN" ]; then
    make_request "Get User Profile" "GET" "/auth/profile" "" "$ACCESS_TOKEN"
    PROFILE_HTTP_STATUS=$?
    if [ "$PROFILE_HTTP_STATUS" -eq 200 ]; then
        print_result "Get User Profile" 0
    else
        PROFILE_MESSAGE=$(jq -r '.message' "$TEST_DIR/Get_User_Profile_response.json" || echo "Unknown error")
        print_result "Get User Profile" 1 "HTTP $PROFILE_HTTP_STATUS: $PROFILE_MESSAGE"
    fi
else
    print_result "Get User Profile" 1 "Skipped: ACCESS_TOKEN not available"
fi

# Test 4: Refresh Token
REFRESH_DATA='{"refresh_token": "'$REFRESH_TOKEN'"}'
make_request "Refresh Token" "POST" "/auth/refresh" "$REFRESH_DATA" ""
REFRESH_HTTP_STATUS=$?
if [ "$REFRESH_HTTP_STATUS" -eq 200 ]; then
    ACCESS_TOKEN=$(jq -r '.access_token' "$TEST_DIR/Refresh_Token_response.json")
    echo "$ACCESS_TOKEN" > "$TEST_DIR/access_token.txt"
    print_result "Refresh Token" 0
else
    REFRESH_MESSAGE=$(jq -r '.message' "$TEST_DIR/Refresh_Token_response.json" || echo "Unknown error")
    print_result "Refresh Token" 1 "HTTP $REFRESH_HTTP_STATUS: $REFRESH_MESSAGE"
fi

# Test 5: Change Password
CHANGE_PASS_DATA='{"current_password": "Password123!", "new_password": "NewPassword456!"}'
make_request "Change Password" "POST" "/auth/change-password" "$CHANGE_PASS_DATA" "$ACCESS_TOKEN"
CHANGE_PASS_HTTP_STATUS=$?
if [ "$CHANGE_PASS_HTTP_STATUS" -eq 200 ]; then
    print_result "Change Password" 0
else
    CHANGE_PASS_MESSAGE=$(jq -r '.message' "$TEST_DIR/Change_Password_response.json" || echo "Unknown error")
    print_result "Change Password" 1 "HTTP $CHANGE_PASS_HTTP_STATUS: $CHANGE_PASS_MESSAGE"
fi

# Test 6: Login with New Password
LOGIN_NEW_PASS_DATA='{"email": "'$USER_EMAIL'", "password": "NewPassword456!"}'
make_request "Login with New Password" "POST" "/auth/login" "$LOGIN_NEW_PASS_DATA" ""
LOGIN_NEW_PASS_HTTP_STATUS=$?
if [ "$LOGIN_NEW_PASS_HTTP_STATUS" -eq 200 ]; then
    ACCESS_TOKEN=$(jq -r '.access_token' "$TEST_DIR/Login_with_New_Password_response.json")
    echo "$ACCESS_TOKEN" > "$TEST_DIR/access_token.txt"
    print_result "Login with New Password" 0
else
    LOGIN_NEW_PASS_MESSAGE=$(jq -r '.message' "$TEST_DIR/Login_with_New_Password_response.json" || echo "Unknown error")
    print_result "Login with New Password" 1 "HTTP $LOGIN_NEW_PASS_HTTP_STATUS: $LOGIN_NEW_PASS_MESSAGE"
fi

# Test 7: Logout
if [ -n "$ACCESS_TOKEN" ]; then
    make_request "Logout User" "POST" "/auth/logout" "" "$ACCESS_TOKEN"
    LOGOUT_HTTP_STATUS=$?
    if [ "$LOGOUT_HTTP_STATUS" -eq 200 ]; then
        print_result "Logout User" 0
    else
        LOGOUT_MESSAGE=$(jq -r '.message' "$TEST_DIR/Logout_User_response.json" || echo "Unknown error")
        print_result "Logout User" 1 "HTTP $LOGOUT_HTTP_STATUS: $LOGOUT_MESSAGE"
    fi
else
    print_result "Logout User" 1 "Skipped: ACCESS_TOKEN not available"
fi

print_header "Summary"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"

# Clean up
rm -f "$TEST_DIR/access_token.txt" "$TEST_DIR/refresh_token.txt"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All authentication tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some authentication tests failed!${NC}"
    exit 1
fi
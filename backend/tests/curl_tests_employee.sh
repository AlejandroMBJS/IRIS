#!/bin/bash

# Employee API Tests with cURL
# =============================

BASE_URL="http://localhost:8080/api/v1"
TOKEN_FILE="./tests/curl_results/access_token.txt" # Adjusted path

# Load token if exists
if [ -f "$TOKEN_FILE" ]; then
    ACCESS_TOKEN=$(cat "$TOKEN_FILE")
else
    echo "Token file not found. Run authentication tests first."
    exit 1
fi

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "\n${GREEN}=== Employee API Tests ===${NC}\n"

# Test 1: List Employees
echo "Test 1: Listing employees..."
response=$(curl -s -w "\n%{http_code}" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "$BASE_URL/employees?page=1&page_size=5")

status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" -eq 200 ]; then
    echo -e "${GREEN}✓ Success${NC}"
    echo "Response preview:"
    echo "$body" | jq '. | {total: .total, page: .page, employees_count: (.employees | length)}'
else
    echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
    echo "$body"
fi

# Test 2: Create Employee
echo -e "\nTest 2: Creating employee..."
EMPLOYEE_RFC="RFC$(date +%s%N)"
EMPLOYEE_CURP="CURP$(date +%s%N)HSPLR09"
employee_data='{
    "employee_number": "TEST-001",
    "first_name": "Test",
    "last_name": "Employee",
    "date_of_birth": "1990-01-01T00:00:00Z",
    "gender": "male",
    "rfc": "'"$EMPLOYEE_RFC"'",
    "curp": "'"$EMPLOYEE_CURP"'",
    "state": "San Luis Potosí",
    "country": "México",
    "hire_date": "2024-01-01T00:00:00Z",
    "employment_status": "active",
    "employee_type": "permanent",
    "daily_salary": 500.00,
    "integrated_daily_salary": 530.00,
    "payment_method": "bank_transfer",
    "regime": "salary"
}'

response=$(curl -s -w "\n%{http_code}" \
  -X POST \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$employee_data" \
  "$BASE_URL/employees")

status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" -eq 201 ]; then
    echo -e "${GREEN}✓ Success${NC}"
    EMPLOYEE_ID=$(echo "$body" | jq -r '.id')
    echo "Created employee ID: $EMPLOYEE_ID"
else
    echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
    echo "$body"
fi

# Test 3: Get Employee Details
if [ -n "$EMPLOYEE_ID" ]; then
    echo -e "\nTest 3: Getting employee details..."
    response=$(curl -s -w "\n%{http_code}" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      "$BASE_URL/employees/$EMPLOYEE_ID")
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "Employee details:"
        echo "$body" | jq '. | {employee_number: .employee_number, full_name: .full_name, daily_salary: .daily_salary}'
    else
        echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
        echo "$body"
    fi
fi

# Test 4: Update Employee
if [ -n "$EMPLOYEE_ID" ]; then
    echo -e "\nTest 4: Updating employee..."
    update_data='{
        "employee_number": "TEST-001",
        "first_name": "Test Updated",
        "last_name": "Employee",
        "date_of_birth": "1990-01-01T00:00:00Z",
        "gender": "male",
        "rfc": "'"$EMPLOYEE_RFC"'",
        "curp": "'"$EMPLOYEE_CURP"'",
        "state": "San Luis Potosí",
        "country": "México",
        "hire_date": "2024-01-01T00:00:00Z",
        "employment_status": "active",
        "employee_type": "permanent",
        "daily_salary": 550.00,
        "integrated_daily_salary": 585.00,
        "payment_method": "bank_transfer",
        "regime": "salary"
    }'
    
    response=$(curl -s -w "\n%{http_code}" \
      -X PUT \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" \
      -d "$update_data" \
      "$BASE_URL/employees/$EMPLOYEE_ID")
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "Updated salary:"
        echo "$body" | jq '.daily_salary'
    else
        echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
        echo "$body"
    fi
fi

# Test 5: Update Salary
if [ -n "$EMPLOYEE_ID" ]; then
    echo -e "\nTest 5: Updating employee salary..."
    salary_data='{
        "new_daily_salary": 600.00,
        "effective_date": "2024-02-01T00:00:00Z",
        "reason": "Annual increase",
        "comments": "Based on performance review"
    }'
    
    response=$(curl -s -w "\n%{http_code}" \
      -X PUT \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" \
      -d "$salary_data" \
      "$BASE_URL/employees/$EMPLOYEE_ID/salary")
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}✓ Success${NC}"
    else
        echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
        echo "$body"
    fi
fi

# Test 6: Terminate Employee
if [ -n "$EMPLOYEE_ID" ]; then
    echo -e "\nTest 6: Terminating employee..."
    termination_data='{
        "termination_date": "2024-12-31T00:00:00Z",
        "reason": "End of contract",
        "comments": "Contract completed"
    }'
    
    response=$(curl -s -w "\n%{http_code}" \
      -X POST \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" \
      -d "$termination_data" \
      "$BASE_URL/employees/$EMPLOYEE_ID/terminate")
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}✓ Success${NC}"
    else
        echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
        echo "$body"
    fi
fi

# Test 7: Get Employee Stats
echo -e "\nTest 7: Getting employee statistics..."
response=$(curl -s -w "\n%{http_code}" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "$BASE_URL/employees/stats")

status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" -eq 200 ]; then
    echo -e "${GREEN}✓ Success${NC}"
    echo "Statistics:"
    echo "$body" | jq '.'
else
    echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
    echo "$body"
fi

# Test 8: Validate Mexican IDs
echo -e "\nTest 8: Validating Mexican IDs..."
response=$(curl -s -w "\n%{http_code}" \
  -X POST \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/employees/validate-ids?rfc=XAXX010101XXX&curp=XAXX010101HDFXXX01&nss=12345678901")

status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$status_code" -eq 200 ]; then
    echo -e "${GREEN}✓ Success${NC}"
    echo "Validation results:"
    echo "$body" | jq '.'
else
    echo -e "${RED}✗ Failed (Status: $status_code)${NC}"
    echo "$body"
fi

echo -e "\n${GREEN}=== Employee API Tests Complete ===${NC}"
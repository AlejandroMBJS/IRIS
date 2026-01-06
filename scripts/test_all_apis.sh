#!/bin/bash
# Comprehensive API Test Script for IRIS Payroll System

BASE_URL="http://localhost:8080/api/v1"
TOKEN=""
EMPLOYEE_ID=""
PERIOD_ID=""
INCIDENCE_TYPE_ID=""
INCIDENCE_ID=""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test result
print_result() {
    local test_name=$1
    local response=$2
    local expected=$3

    if echo "$response" | grep -q "$expected"; then
        echo -e "${GREEN}[PASS]${NC} $test_name"
        return 0
    else
        echo -e "${RED}[FAIL]${NC} $test_name"
        echo "Response: $response"
        return 1
    fi
}

echo "================================================"
echo "IRIS Payroll System - Comprehensive API Tests"
echo "================================================"

# ==========================================
# HEALTH CHECK TESTS
# ==========================================
echo -e "\n${YELLOW}=== Health Check Tests ===${NC}"

# Test 1: Health endpoint
RESPONSE=$(curl -s "$BASE_URL/health")
print_result "GET /health" "$RESPONSE" "healthy"

# ==========================================
# AUTHENTICATION TESTS
# ==========================================
echo -e "\n${YELLOW}=== Authentication Tests ===${NC}"

# Test 2: Register new company
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "company_name": "Test Company API",
        "company_rfc": "TCO123456ABC",
        "email": "apitest@company.com",
        "password": "TestPass123@",
        "role": "admin",
        "full_name": "API Test Admin"
    }')
print_result "POST /auth/register" "$RESPONSE" "access_token"

# Test 3: Login
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email": "apitest@company.com", "password": "TestPass123@"}')
print_result "POST /auth/login" "$RESPONSE" "access_token"

# Extract token for authenticated requests
TOKEN=$(echo $RESPONSE | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"$//')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to get authentication token. Aborting tests.${NC}"
    exit 1
fi

echo "Token acquired successfully"

# Test 4: Get profile
RESPONSE=$(curl -s -X GET "$BASE_URL/auth/profile" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /auth/profile" "$RESPONSE" "apitest@company.com"

# Test 5: Update profile
RESPONSE=$(curl -s -X PUT "$BASE_URL/auth/profile" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"full_name": "Updated Admin Name"}')
print_result "PUT /auth/profile" "$RESPONSE" "Updated Admin Name"

# Test 6: Forgot password
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/forgot-password" \
    -H "Content-Type: application/json" \
    -d '{"email": "apitest@company.com"}')
print_result "POST /auth/forgot-password" "$RESPONSE" "password reset"

# ==========================================
# EMPLOYEE TESTS
# ==========================================
echo -e "\n${YELLOW}=== Employee Tests ===${NC}"

# Test 7: Create employee
RESPONSE=$(curl -s -X POST "$BASE_URL/employees" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "employee_number": "EMP001",
        "first_name": "Juan",
        "last_name": "Perez",
        "date_of_birth": "1990-05-15T00:00:00Z",
        "gender": "male",
        "rfc": "PERJ900515XXX",
        "curp": "PERJ900515HSLRLN09",
        "hire_date": "2024-01-15T00:00:00Z",
        "daily_salary": 500.00,
        "employment_status": "active",
        "employee_type": "permanent",
        "collar_type": "white_collar",
        "pay_frequency": "weekly",
        "payment_method": "transfer"
    }')
print_result "POST /employees" "$RESPONSE" "EMP001"

EMPLOYEE_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"$//')

if [ -z "$EMPLOYEE_ID" ]; then
    echo -e "${RED}Failed to create employee. Using placeholder.${NC}"
    EMPLOYEE_ID="00000000-0000-0000-0000-000000000000"
fi

# Test 8: Get all employees
RESPONSE=$(curl -s -X GET "$BASE_URL/employees" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /employees" "$RESPONSE" "EMP001"

# Test 9: Get employee by ID
RESPONSE=$(curl -s -X GET "$BASE_URL/employees/$EMPLOYEE_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /employees/:id" "$RESPONSE" "Juan"

# Test 10: Update employee
RESPONSE=$(curl -s -X PUT "$BASE_URL/employees/$EMPLOYEE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "employee_number": "EMP001",
        "first_name": "Juan Carlos",
        "last_name": "Perez",
        "date_of_birth": "1990-05-15T00:00:00Z",
        "gender": "male",
        "rfc": "PERJ900515XXX",
        "curp": "PERJ900515HSLRLN09",
        "hire_date": "2024-01-15T00:00:00Z",
        "daily_salary": 550.00,
        "employment_status": "active",
        "employee_type": "permanent",
        "collar_type": "white_collar",
        "pay_frequency": "weekly",
        "payment_method": "transfer"
    }')
print_result "PUT /employees/:id" "$RESPONSE" "Juan Carlos"

# Test 11: Get employee stats
RESPONSE=$(curl -s -X GET "$BASE_URL/employees/stats" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /employees/stats" "$RESPONSE" "total_employees"

# Test 12: Update salary
RESPONSE=$(curl -s -X PUT "$BASE_URL/employees/$EMPLOYEE_ID/salary" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"new_daily_salary": 600.00, "effective_date": "2025-01-01"}')
print_result "PUT /employees/:id/salary" "$RESPONSE" "600"

# ==========================================
# PAYROLL PERIOD TESTS
# ==========================================
echo -e "\n${YELLOW}=== Payroll Period Tests ===${NC}"

# Test 13: Create payroll period
RESPONSE=$(curl -s -X POST "$BASE_URL/payroll/periods" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "year": 2025,
        "period_number": 1,
        "start_date": "2025-01-01",
        "end_date": "2025-01-07",
        "payment_date": "2025-01-10",
        "frequency": "weekly"
    }')
print_result "POST /payroll/periods" "$RESPONSE" "2025"

PERIOD_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"$//')

if [ -z "$PERIOD_ID" ]; then
    echo -e "${RED}Failed to create period. Using placeholder.${NC}"
    PERIOD_ID="00000000-0000-0000-0000-000000000000"
fi

# Test 14: Get all periods
RESPONSE=$(curl -s -X GET "$BASE_URL/payroll/periods" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /payroll/periods" "$RESPONSE" "2025"

# Test 15: Get period by ID
RESPONSE=$(curl -s -X GET "$BASE_URL/payroll/periods/$PERIOD_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /payroll/periods/:id" "$RESPONSE" "weekly"

# ==========================================
# INCIDENCE TYPE TESTS
# ==========================================
echo -e "\n${YELLOW}=== Incidence Type Tests ===${NC}"

# Test 16: Create incidence type
RESPONSE=$(curl -s -X POST "$BASE_URL/incidence-types" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Falta Injustificada",
        "category": "absence",
        "effect_type": "negative",
        "is_calculated": true,
        "calculation_method": "daily_rate",
        "default_value": 0,
        "description": "Falta sin justificacion"
    }')
print_result "POST /incidence-types" "$RESPONSE" "Falta Injustificada"

INCIDENCE_TYPE_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"$//')

# Test 17: Get all incidence types
RESPONSE=$(curl -s -X GET "$BASE_URL/incidence-types" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /incidence-types" "$RESPONSE" "Falta Injustificada"

# Test 18: Update incidence type
RESPONSE=$(curl -s -X PUT "$BASE_URL/incidence-types/$INCIDENCE_TYPE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Falta Injustificada Updated",
        "category": "absence",
        "effect_type": "negative",
        "is_calculated": true,
        "calculation_method": "daily_rate",
        "default_value": 0,
        "description": "Updated description"
    }')
print_result "PUT /incidence-types/:id" "$RESPONSE" "Updated"

# ==========================================
# INCIDENCE TESTS
# ==========================================
echo -e "\n${YELLOW}=== Incidence Tests ===${NC}"

# Test 19: Create incidence
RESPONSE=$(curl -s -X POST "$BASE_URL/incidences" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"employee_id\": \"$EMPLOYEE_ID\",
        \"payroll_period_id\": \"$PERIOD_ID\",
        \"incidence_type_id\": \"$INCIDENCE_TYPE_ID\",
        \"start_date\": \"2025-01-02\",
        \"end_date\": \"2025-01-02\",
        \"quantity\": 1,
        \"comments\": \"Test incidence\"
    }")
print_result "POST /incidences" "$RESPONSE" "pending"

INCIDENCE_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"$//')

# Test 20: Get all incidences
RESPONSE=$(curl -s -X GET "$BASE_URL/incidences" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /incidences" "$RESPONSE" "incidence"

# Test 21: Get incidence by ID
RESPONSE=$(curl -s -X GET "$BASE_URL/incidences/$INCIDENCE_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /incidences/:id" "$RESPONSE" "Test incidence"

# Test 22: Update incidence
RESPONSE=$(curl -s -X PUT "$BASE_URL/incidences/$INCIDENCE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"comments": "Updated comment"}')
print_result "PUT /incidences/:id" "$RESPONSE" "Updated comment"

# Test 23: Approve incidence
RESPONSE=$(curl -s -X POST "$BASE_URL/incidences/$INCIDENCE_ID/approve" \
    -H "Authorization: Bearer $TOKEN")
print_result "POST /incidences/:id/approve" "$RESPONSE" "approved"

# Test 24: Get employee incidences
RESPONSE=$(curl -s -X GET "$BASE_URL/employees/$EMPLOYEE_ID/incidences" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /employees/:id/incidences" "$RESPONSE" "incidence"

# Test 25: Get vacation balance
RESPONSE=$(curl -s -X GET "$BASE_URL/employees/$EMPLOYEE_ID/vacation-balance" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /employees/:id/vacation-balance" "$RESPONSE" "entitled_days"

# Test 26: Get absence summary
RESPONSE=$(curl -s -X GET "$BASE_URL/employees/$EMPLOYEE_ID/absence-summary?year=2025" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /employees/:id/absence-summary" "$RESPONSE" "employee_id"

# ==========================================
# PAYROLL CALCULATION TESTS
# ==========================================
echo -e "\n${YELLOW}=== Payroll Calculation Tests ===${NC}"

# Test 27: Calculate payroll
RESPONSE=$(curl -s -X POST "$BASE_URL/payroll/calculate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"employee_id\": \"$EMPLOYEE_ID\", \"payroll_period_id\": \"$PERIOD_ID\"}")
print_result "POST /payroll/calculate" "$RESPONSE" "total_net_pay"

# Test 28: Get payroll calculation
RESPONSE=$(curl -s -X GET "$BASE_URL/payroll/calculation/$PERIOD_ID/$EMPLOYEE_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /payroll/calculation/:period/:employee" "$RESPONSE" "net_pay"

# Test 29: Get payroll by period
RESPONSE=$(curl -s -X GET "$BASE_URL/payroll/period/$PERIOD_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /payroll/period/:id" "$RESPONSE" "payroll"

# Test 30: Get payroll summary
RESPONSE=$(curl -s -X GET "$BASE_URL/payroll/summary/$PERIOD_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /payroll/summary/:id" "$RESPONSE" "total"

# Test 31: Get concept totals
RESPONSE=$(curl -s -X GET "$BASE_URL/payroll/concept-totals/$PERIOD_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /payroll/concept-totals/:id" "$RESPONSE" "["

# Test 32: Get payslip (PDF)
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/payroll/payslip/$PERIOD_ID/$EMPLOYEE_ID?format=pdf" \
    -H "Authorization: Bearer $TOKEN")
if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}[PASS]${NC} GET /payroll/payslip/:period/:employee?format=pdf"
else
    echo -e "${RED}[FAIL]${NC} GET /payroll/payslip/:period/:employee?format=pdf (HTTP $RESPONSE)"
fi

# Test 33: Get payslip (XML)
RESPONSE=$(curl -s "$BASE_URL/payroll/payslip/$PERIOD_ID/$EMPLOYEE_ID?format=xml" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /payroll/payslip/:period/:employee?format=xml" "$RESPONSE" "xml"

# ==========================================
# CATALOG TESTS
# ==========================================
echo -e "\n${YELLOW}=== Catalog Tests ===${NC}"

# Test 34: Get payroll concepts
RESPONSE=$(curl -s -X GET "$BASE_URL/catalogs/concepts" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /catalogs/concepts" "$RESPONSE" "["

# Test 35: Get catalog incidence types
RESPONSE=$(curl -s -X GET "$BASE_URL/catalogs/incidence-types" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /catalogs/incidence-types" "$RESPONSE" "["

# ==========================================
# REPORT TESTS
# ==========================================
echo -e "\n${YELLOW}=== Report Tests ===${NC}"

# Test 36: Generate report
RESPONSE=$(curl -s -X POST "$BASE_URL/reports/generate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"report_type\": \"payroll_summary\", \"payroll_period_id\": \"$PERIOD_ID\", \"format\": \"json\"}")
print_result "POST /reports/generate" "$RESPONSE" "report"

# Test 37: Get report history
RESPONSE=$(curl -s -X GET "$BASE_URL/reports/history" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /reports/history" "$RESPONSE" "["

# ==========================================
# USER MANAGEMENT TESTS (Admin Only)
# ==========================================
echo -e "\n${YELLOW}=== User Management Tests ===${NC}"

# Test 38: Get users
RESPONSE=$(curl -s -X GET "$BASE_URL/users" \
    -H "Authorization: Bearer $TOKEN")
print_result "GET /users" "$RESPONSE" "apitest@company.com"

# Test 39: Create user
RESPONSE=$(curl -s -X POST "$BASE_URL/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "staff@company.com",
        "password": "StaffPass123@",
        "full_name": "Staff Member",
        "role": "payroll_staff"
    }')
print_result "POST /users" "$RESPONSE" "staff@company.com"

NEW_USER_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"$//')

# Test 40: Toggle user active
RESPONSE=$(curl -s -X PATCH "$BASE_URL/users/$NEW_USER_ID/toggle-active" \
    -H "Authorization: Bearer $TOKEN")
print_result "PATCH /users/:id/toggle-active" "$RESPONSE" "is_active"

# Test 41: Delete user
RESPONSE=$(curl -s -X DELETE "$BASE_URL/users/$NEW_USER_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "DELETE /users/:id" "$RESPONSE" "deleted"

# ==========================================
# CLEANUP TESTS (Delete Operations)
# ==========================================
echo -e "\n${YELLOW}=== Cleanup Tests ===${NC}"

# Create new incidence for delete test
RESPONSE=$(curl -s -X POST "$BASE_URL/incidences" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"employee_id\": \"$EMPLOYEE_ID\",
        \"payroll_period_id\": \"$PERIOD_ID\",
        \"incidence_type_id\": \"$INCIDENCE_TYPE_ID\",
        \"start_date\": \"2025-01-03\",
        \"end_date\": \"2025-01-03\",
        \"quantity\": 1,
        \"comments\": \"To be deleted\"
    }")
DELETE_INCIDENCE_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"$//')

# Test 42: Delete incidence
RESPONSE=$(curl -s -X DELETE "$BASE_URL/incidences/$DELETE_INCIDENCE_ID" \
    -H "Authorization: Bearer $TOKEN")
print_result "DELETE /incidences/:id" "$RESPONSE" "deleted"

# Test 43: Logout
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
    -H "Authorization: Bearer $TOKEN")
print_result "POST /auth/logout" "$RESPONSE" "logged out"

echo ""
echo "================================================"
echo "API Tests Complete!"
echo "================================================"

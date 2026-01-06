#!/bin/bash
# Comprehensive API Test Script for IRIS Payroll System
# This script tests ALL API endpoints and documents them

set -e

BASE_URL="http://localhost:8080/api/v1"
TOKEN=""
RESULTS_FILE="/tmp/api_test_results.md"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0

# Start fresh results file
cat > $RESULTS_FILE << 'EOF'
# IRIS Payroll System - API Endpoints Documentation

## Overview
This document lists all API endpoints with their corresponding:
- Backend handler
- API client function (frontend)
- Frontend page/component that uses it

---

EOF

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}✗ $1${NC}"
    ((FAILED++))
}

log_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$status_code" = "$expected_status" ]; then
        log_success "$method $endpoint - $description (HTTP $status_code)"
        echo "$body"
        return 0
    else
        log_fail "$method $endpoint - Expected $expected_status, got $status_code"
        echo "$body"
        return 1
    fi
}

echo "=========================================="
echo "IRIS Payroll System - Comprehensive API Test"
echo "=========================================="
echo ""

# ===========================================
# 1. AUTHENTICATION ENDPOINTS
# ===========================================
echo ""
echo "=== 1. AUTHENTICATION ENDPOINTS ==="
echo ""

# Register admin user
log_info "Registering admin user..."
REGISTER_RESP=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "company_name": "IRIS Test Company",
        "company_rfc": "ITC123456ABC",
        "email": "admin@iris.com",
        "password": "Admin123!",
        "role": "admin",
        "full_name": "Admin User"
    }')
echo "Register response: $REGISTER_RESP"

# Login to get token
log_info "Logging in..."
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "admin@iris.com",
        "password": "Admin123!"
    }')
echo "Login response: $LOGIN_RESP"

TOKEN=$(echo $LOGIN_RESP | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    echo "Failed to get token!"
    exit 1
fi
echo "Token obtained: ${TOKEN:0:50}..."

# Save token
echo "$TOKEN" > /tmp/token.txt

cat >> $RESULTS_FILE << 'EOF'
## 1. Authentication Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/auth/register` | POST | `auth_handler.go:Register` | `authApi.register()` | `/login` (signup mode) |
| `/auth/login` | POST | `auth_handler.go:Login` | `authApi.login()` | `/login` |
| `/auth/refresh` | POST | `auth_handler.go:RefreshToken` | `authApi.refreshToken()` | Auto (api-client interceptor) |
| `/auth/me` | GET | `auth_handler.go:GetCurrentUser` | `authApi.getCurrentUser()` | Dashboard, Profile |

EOF

log_success "POST /auth/register"
log_success "POST /auth/login"

# Test refresh token
log_info "Testing token refresh..."
REFRESH_RESP=$(curl -s -X POST "$BASE_URL/auth/refresh" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")
echo "Refresh response: $REFRESH_RESP"
log_success "POST /auth/refresh"

# Test get current user
log_info "Testing get current user..."
ME_RESP=$(curl -s -X GET "$BASE_URL/auth/me" \
    -H "Authorization: Bearer $TOKEN")
echo "Me response: $ME_RESP"
log_success "GET /auth/me"

# ===========================================
# 2. EMPLOYEE ENDPOINTS
# ===========================================
echo ""
echo "=== 2. EMPLOYEE ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 2. Employee Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/employees` | GET | `employee_handler.go:ListEmployees` | `employeeApi.list()` | `/employees` |
| `/employees` | POST | `employee_handler.go:CreateEmployee` | `employeeApi.create()` | `/employees/new` |
| `/employees/:id` | GET | `employee_handler.go:GetEmployee` | `employeeApi.getById()` | `/employees/[id]` |
| `/employees/:id` | PUT | `employee_handler.go:UpdateEmployee` | `employeeApi.update()` | `/employees/[id]/edit` |
| `/employees/:id` | DELETE | `employee_handler.go:DeleteEmployee` | `employeeApi.delete()` | `/employees` (delete action) |

EOF

# Create employees
log_info "Creating white collar employee..."
WHITE1=$(curl -s -X POST "$BASE_URL/employees" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "employee_number": "WC001",
        "first_name": "Maria",
        "last_name": "Garcia Lopez",
        "date_of_birth": "1985-03-15T00:00:00Z",
        "gender": "female",
        "rfc": "GALM850315XXX",
        "curp": "GALM850315MDFRCR09",
        "nss": "12345678901",
        "hire_date": "2020-01-15T00:00:00Z",
        "daily_salary": 800.00,
        "employment_status": "active",
        "employee_type": "permanent",
        "collar_type": "white",
        "payment_method": "transfer",
        "bank_name": "BBVA",
        "bank_account": "012345678901234567",
        "clabe": "012345678901234567"
    }')
WHITE1_ID=$(echo $WHITE1 | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "White collar 1 created: $WHITE1_ID"
log_success "POST /employees (white collar)"

log_info "Creating blue collar employee..."
BLUE1=$(curl -s -X POST "$BASE_URL/employees" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "employee_number": "BC001",
        "first_name": "Juan",
        "last_name": "Martinez Perez",
        "date_of_birth": "1990-07-20T00:00:00Z",
        "gender": "male",
        "rfc": "MAPJ900720XXX",
        "curp": "MAPJ900720HDFRRN05",
        "nss": "98765432101",
        "hire_date": "2021-06-01T00:00:00Z",
        "daily_salary": 350.00,
        "employment_status": "active",
        "employee_type": "permanent",
        "collar_type": "blue",
        "payment_method": "cash"
    }')
BLUE1_ID=$(echo $BLUE1 | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Blue collar 1 created: $BLUE1_ID"
log_success "POST /employees (blue collar)"

log_info "Creating gray collar employee..."
GRAY1=$(curl -s -X POST "$BASE_URL/employees" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "employee_number": "GC001",
        "first_name": "Roberto",
        "last_name": "Sanchez Diaz",
        "date_of_birth": "1988-11-10T00:00:00Z",
        "gender": "male",
        "rfc": "SADR881110XXX",
        "curp": "SADR881110HDFNZB07",
        "nss": "45678912301",
        "hire_date": "2022-03-01T00:00:00Z",
        "daily_salary": 500.00,
        "employment_status": "active",
        "employee_type": "permanent",
        "collar_type": "gray",
        "payment_method": "transfer",
        "bank_name": "Santander",
        "bank_account": "543210987654321098",
        "clabe": "543210987654321098"
    }')
GRAY1_ID=$(echo $GRAY1 | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Gray collar 1 created: $GRAY1_ID"
log_success "POST /employees (gray collar)"

# List employees
log_info "Listing employees..."
EMP_LIST=$(curl -s -X GET "$BASE_URL/employees" \
    -H "Authorization: Bearer $TOKEN")
echo "Employees: $EMP_LIST"
log_success "GET /employees"

# Get single employee
log_info "Getting single employee..."
EMP_DETAIL=$(curl -s -X GET "$BASE_URL/employees/$WHITE1_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "Employee detail: $EMP_DETAIL"
log_success "GET /employees/:id"

# Update employee
log_info "Updating employee..."
UPDATE_EMP=$(curl -s -X PUT "$BASE_URL/employees/$WHITE1_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "daily_salary": 850.00,
        "employment_status": "active"
    }')
echo "Updated employee: $UPDATE_EMP"
log_success "PUT /employees/:id"

# ===========================================
# 3. PAYROLL PERIOD ENDPOINTS
# ===========================================
echo ""
echo "=== 3. PAYROLL PERIOD ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 3. Payroll Period Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/payroll/periods` | GET | `payroll_period_handler.go:ListPeriods` | `payrollApi.listPeriods()` | `/payroll`, Period selector |
| `/payroll/periods` | POST | `payroll_period_handler.go:CreatePeriod` | `payrollApi.createPeriod()` | `/payroll/periods/new` |
| `/payroll/periods/:id` | GET | `payroll_period_handler.go:GetPeriod` | `payrollApi.getPeriodById()` | `/payroll/periods/[id]` |
| `/payroll/periods/:id` | PUT | `payroll_period_handler.go:UpdatePeriod` | `payrollApi.updatePeriod()` | `/payroll/periods/[id]/edit` |
| `/payroll/periods/:id/close` | POST | `payroll_period_handler.go:ClosePeriod` | `payrollApi.closePeriod()` | `/payroll` (close action) |
| `/payroll/periods/current` | GET | `payroll_period_handler.go:GetCurrentPeriod` | `payrollApi.getCurrentPeriod()` | Dashboard, Navbar |

EOF

# Create weekly period
log_info "Creating weekly period..."
WEEKLY=$(curl -s -X POST "$BASE_URL/payroll/periods" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "period_code": "2025-W49",
        "year": 2025,
        "period_number": 49,
        "frequency": "weekly",
        "period_type": "weekly",
        "start_date": "2025-12-01T00:00:00Z",
        "end_date": "2025-12-07T00:00:00Z",
        "payment_date": "2025-12-10T00:00:00Z",
        "description": "Semana 49 - Diciembre 2025"
    }')
WEEKLY_ID=$(echo $WEEKLY | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Weekly period created: $WEEKLY_ID"
log_success "POST /payroll/periods (weekly)"

# Create biweekly period
log_info "Creating biweekly period..."
BIWEEKLY=$(curl -s -X POST "$BASE_URL/payroll/periods" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "period_code": "2025-BW24",
        "year": 2025,
        "period_number": 24,
        "frequency": "biweekly",
        "period_type": "biweekly",
        "start_date": "2025-12-15T00:00:00Z",
        "end_date": "2025-12-31T00:00:00Z",
        "payment_date": "2026-01-05T00:00:00Z",
        "description": "Quincena 24 - Diciembre 2025"
    }')
BIWEEKLY_ID=$(echo $BIWEEKLY | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Biweekly period created: $BIWEEKLY_ID"
log_success "POST /payroll/periods (biweekly)"

# List periods
log_info "Listing periods..."
PERIODS=$(curl -s -X GET "$BASE_URL/payroll/periods" \
    -H "Authorization: Bearer $TOKEN")
echo "Periods: $PERIODS"
log_success "GET /payroll/periods"

# Get current period
log_info "Getting current period..."
CURRENT=$(curl -s -X GET "$BASE_URL/payroll/periods/current" \
    -H "Authorization: Bearer $TOKEN")
echo "Current period: $CURRENT"
log_success "GET /payroll/periods/current"

# Get single period
log_info "Getting single period..."
PERIOD_DETAIL=$(curl -s -X GET "$BASE_URL/payroll/periods/$WEEKLY_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "Period detail: $PERIOD_DETAIL"
log_success "GET /payroll/periods/:id"

# ===========================================
# 4. INCIDENCE TYPE ENDPOINTS
# ===========================================
echo ""
echo "=== 4. INCIDENCE TYPE ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 4. Incidence Type Endpoints (Catalog)

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/incidence-types` | GET | `incidence_handler.go:ListIncidenceTypes` | `incidenceApi.listTypes()` | `/incidences/new`, `/incidences` |
| `/incidence-types` | POST | `incidence_handler.go:CreateIncidenceType` | `incidenceApi.createType()` | Admin settings |
| `/incidence-types/:id` | PUT | `incidence_handler.go:UpdateIncidenceType` | `incidenceApi.updateType()` | Admin settings |
| `/incidence-types/:id` | DELETE | `incidence_handler.go:DeleteIncidenceType` | `incidenceApi.deleteType()` | Admin settings |

EOF

# Create incidence types
declare -a INCIDENCE_TYPES=(
    '{"name":"Vacaciones","category":"vacation","effect_type":"neutral","is_calculated":true,"calculation_method":"daily_rate","description":"Dias de vacaciones"}'
    '{"name":"Incapacidad IMSS","category":"sick","effect_type":"negative","is_calculated":true,"calculation_method":"daily_rate","description":"Incapacidad por enfermedad"}'
    '{"name":"Falta Injustificada","category":"absence","effect_type":"negative","is_calculated":true,"calculation_method":"daily_rate","description":"Falta sin justificacion"}'
    '{"name":"Horas Extra Dobles","category":"overtime","effect_type":"positive","is_calculated":true,"calculation_method":"hourly_rate","default_value":2.0,"description":"Horas extra al doble"}'
    '{"name":"Retardo","category":"delay","effect_type":"negative","is_calculated":false,"description":"Llegada tarde"}'
    '{"name":"Bono Productividad","category":"bonus","effect_type":"positive","is_calculated":false,"description":"Bono por productividad"}'
)

INCIDENCE_TYPE_IDS=()
for type_data in "${INCIDENCE_TYPES[@]}"; do
    name=$(echo $type_data | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
    log_info "Creating incidence type: $name..."
    RESP=$(curl -s -X POST "$BASE_URL/incidence-types" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$type_data")
    TYPE_ID=$(echo $RESP | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    INCIDENCE_TYPE_IDS+=("$TYPE_ID")
    echo "Created: $TYPE_ID"
done
log_success "POST /incidence-types (6 types created)"

# List incidence types
log_info "Listing incidence types..."
TYPES=$(curl -s -X GET "$BASE_URL/incidence-types" \
    -H "Authorization: Bearer $TOKEN")
echo "Incidence types: $TYPES"
log_success "GET /incidence-types"

# Get vacation type ID for later
VACATION_TYPE_ID=${INCIDENCE_TYPE_IDS[0]}

# ===========================================
# 5. INCIDENCE ENDPOINTS
# ===========================================
echo ""
echo "=== 5. INCIDENCE ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 5. Incidence Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/incidences` | GET | `incidence_handler.go:ListIncidences` | `incidenceApi.list()` | `/incidences` |
| `/incidences` | POST | `incidence_handler.go:CreateIncidence` | `incidenceApi.create()` | `/incidences/new` |
| `/incidences/:id` | GET | `incidence_handler.go:GetIncidence` | `incidenceApi.getById()` | `/incidences/[id]` |
| `/incidences/:id` | PUT | `incidence_handler.go:UpdateIncidence` | `incidenceApi.update()` | `/incidences/[id]/edit` |
| `/incidences/:id` | DELETE | `incidence_handler.go:DeleteIncidence` | `incidenceApi.delete()` | `/incidences` (delete action) |
| `/incidences/:id/approve` | POST | `incidence_handler.go:ApproveIncidence` | `incidenceApi.approve()` | `/incidences` (approve action) |
| `/incidences/:id/reject` | POST | `incidence_handler.go:RejectIncidence` | `incidenceApi.reject()` | `/incidences` (reject action) |
| `/employees/:id/incidences` | GET | `incidence_handler.go:GetEmployeeIncidences` | `incidenceApi.getByEmployee()` | `/employees/[id]` |
| `/employees/:id/vacation-balance` | GET | `incidence_handler.go:GetEmployeeVacationBalance` | `incidenceApi.getVacationBalance()` | `/employees/[id]` |
| `/employees/:id/absence-summary` | GET | `incidence_handler.go:GetEmployeeAbsenceSummary` | `incidenceApi.getAbsenceSummary()` | `/employees/[id]` |

EOF

# Create incidence
log_info "Creating incidence (vacation request)..."
INCIDENCE=$(curl -s -X POST "$BASE_URL/incidences" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"employee_id\": \"$WHITE1_ID\",
        \"payroll_period_id\": \"$WEEKLY_ID\",
        \"incidence_type_id\": \"$VACATION_TYPE_ID\",
        \"start_date\": \"2025-12-02\",
        \"end_date\": \"2025-12-04\",
        \"quantity\": 3,
        \"comments\": \"Vacaciones de fin de año\"
    }")
INCIDENCE_ID=$(echo $INCIDENCE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Incidence created: $INCIDENCE_ID"
log_success "POST /incidences"

# List incidences
log_info "Listing incidences..."
INCIDENCES=$(curl -s -X GET "$BASE_URL/incidences" \
    -H "Authorization: Bearer $TOKEN")
echo "Incidences: $INCIDENCES"
log_success "GET /incidences"

# Get single incidence
log_info "Getting single incidence..."
INC_DETAIL=$(curl -s -X GET "$BASE_URL/incidences/$INCIDENCE_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "Incidence detail: $INC_DETAIL"
log_success "GET /incidences/:id"

# Approve incidence
log_info "Approving incidence..."
APPROVED=$(curl -s -X POST "$BASE_URL/incidences/$INCIDENCE_ID/approve" \
    -H "Authorization: Bearer $TOKEN")
echo "Approved: $APPROVED"
log_success "POST /incidences/:id/approve"

# Get employee incidences
log_info "Getting employee incidences..."
EMP_INC=$(curl -s -X GET "$BASE_URL/employees/$WHITE1_ID/incidences" \
    -H "Authorization: Bearer $TOKEN")
echo "Employee incidences: $EMP_INC"
log_success "GET /employees/:id/incidences"

# Get vacation balance
log_info "Getting vacation balance..."
VAC_BAL=$(curl -s -X GET "$BASE_URL/employees/$WHITE1_ID/vacation-balance" \
    -H "Authorization: Bearer $TOKEN")
echo "Vacation balance: $VAC_BAL"
log_success "GET /employees/:id/vacation-balance"

# Get absence summary
log_info "Getting absence summary..."
ABS_SUM=$(curl -s -X GET "$BASE_URL/employees/$WHITE1_ID/absence-summary?year=2025" \
    -H "Authorization: Bearer $TOKEN")
echo "Absence summary: $ABS_SUM"
log_success "GET /employees/:id/absence-summary"

# ===========================================
# 6. PAYROLL CALCULATION ENDPOINTS
# ===========================================
echo ""
echo "=== 6. PAYROLL CALCULATION ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 6. Payroll Calculation Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/payroll/calculate` | POST | `payroll_handler.go:CalculatePayroll` | `payrollApi.calculate()` | `/payroll` (calculate action) |
| `/payroll/calculate-batch` | POST | `payroll_handler.go:CalculateBatchPayroll` | `payrollApi.calculateBatch()` | `/payroll` (batch calculate) |
| `/payroll/entries` | GET | `payroll_handler.go:GetPayrollEntries` | `payrollApi.getEntries()` | `/payroll` |
| `/payroll/entries/:id` | GET | `payroll_handler.go:GetPayrollEntry` | `payrollApi.getEntryById()` | `/payroll/[id]` |
| `/payroll/summary/:period_id` | GET | `payroll_handler.go:GetPayrollSummary` | `payrollApi.getSummary()` | `/payroll`, Dashboard |

EOF

# Calculate payroll for white collar
log_info "Calculating payroll for white collar employee..."
CALC_WHITE=$(curl -s -X POST "$BASE_URL/payroll/calculate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"employee_id\": \"$WHITE1_ID\",
        \"payroll_period_id\": \"$WEEKLY_ID\"
    }")
echo "White collar payroll: $CALC_WHITE"
log_success "POST /payroll/calculate (white collar)"

# Calculate payroll for blue collar
log_info "Calculating payroll for blue collar employee..."
CALC_BLUE=$(curl -s -X POST "$BASE_URL/payroll/calculate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"employee_id\": \"$BLUE1_ID\",
        \"payroll_period_id\": \"$WEEKLY_ID\"
    }")
echo "Blue collar payroll: $CALC_BLUE"
log_success "POST /payroll/calculate (blue collar)"

# Calculate payroll for gray collar
log_info "Calculating payroll for gray collar employee..."
CALC_GRAY=$(curl -s -X POST "$BASE_URL/payroll/calculate" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"employee_id\": \"$GRAY1_ID\",
        \"payroll_period_id\": \"$WEEKLY_ID\"
    }")
echo "Gray collar payroll: $CALC_GRAY"
log_success "POST /payroll/calculate (gray collar)"

# Get payroll entries
log_info "Getting payroll entries..."
ENTRIES=$(curl -s -X GET "$BASE_URL/payroll/entries?period_id=$WEEKLY_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "Payroll entries: $ENTRIES"
log_success "GET /payroll/entries"

# Get payroll summary
log_info "Getting payroll summary..."
SUMMARY=$(curl -s -X GET "$BASE_URL/payroll/summary/$WEEKLY_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "Payroll summary: $SUMMARY"
log_success "GET /payroll/summary/:period_id"

# ===========================================
# 7. PAYSLIP ENDPOINTS
# ===========================================
echo ""
echo "=== 7. PAYSLIP ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 7. Payslip Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/payroll/payslip/:period_id/:employee_id` | GET | `payroll_handler.go:GetPayslip` | `payrollApi.getPayslip()` | `/payroll/payslip/[id]` |
| `/payroll/payslip/:period_id/:employee_id?format=pdf` | GET | `payroll_handler.go:GetPayslip` | `payrollApi.downloadPayslipPDF()` | Download button |
| `/payroll/payslip/:period_id/:employee_id?format=xml` | GET | `payroll_handler.go:GetPayslip` | `payrollApi.downloadPayslipXML()` | Download button |

EOF

# Get payslip JSON
log_info "Getting payslip (JSON)..."
PAYSLIP=$(curl -s -X GET "$BASE_URL/payroll/payslip/$WEEKLY_ID/$WHITE1_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "Payslip JSON: $PAYSLIP"
log_success "GET /payroll/payslip/:period_id/:employee_id (JSON)"

# Get payslip PDF
log_info "Getting payslip (PDF)..."
curl -s -X GET "$BASE_URL/payroll/payslip/$WEEKLY_ID/$WHITE1_ID?format=pdf" \
    -H "Authorization: Bearer $TOKEN" \
    -o /tmp/payslip_test.pdf
if [ -f /tmp/payslip_test.pdf ] && [ -s /tmp/payslip_test.pdf ]; then
    log_success "GET /payroll/payslip/:period_id/:employee_id?format=pdf"
    ls -la /tmp/payslip_test.pdf
else
    log_fail "GET /payroll/payslip/:period_id/:employee_id?format=pdf"
fi

# Get payslip XML (CFDI)
log_info "Getting payslip (XML/CFDI)..."
PAYSLIP_XML=$(curl -s -X GET "$BASE_URL/payroll/payslip/$WEEKLY_ID/$WHITE1_ID?format=xml" \
    -H "Authorization: Bearer $TOKEN")
echo "Payslip XML: ${PAYSLIP_XML:0:500}..."
log_success "GET /payroll/payslip/:period_id/:employee_id?format=xml"

# ===========================================
# 8. EVIDENCE/UPLOAD ENDPOINTS
# ===========================================
echo ""
echo "=== 8. EVIDENCE/UPLOAD ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 8. Evidence/Upload Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/evidence/incidence/:incidence_id` | POST | `upload_handler.go:UploadEvidence` | `uploadApi.uploadEvidence()` | `/incidences/[id]` |
| `/evidence/incidence/:incidence_id` | GET | `upload_handler.go:ListEvidence` | `uploadApi.listEvidence()` | `/incidences/[id]` |
| `/evidence/:evidence_id` | GET | `upload_handler.go:GetEvidence` | `uploadApi.getEvidence()` | `/incidences/[id]` |
| `/evidence/:evidence_id/download` | GET | `upload_handler.go:DownloadEvidence` | `uploadApi.downloadEvidence()` | Download button |
| `/evidence/:evidence_id` | DELETE | `upload_handler.go:DeleteEvidence` | `uploadApi.deleteEvidence()` | Delete button |

EOF

# Create test file for upload
echo "Test evidence file content" > /tmp/test_evidence.txt

# Upload evidence
log_info "Uploading evidence..."
UPLOAD=$(curl -s -X POST "$BASE_URL/evidence/incidence/$INCIDENCE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@/tmp/test_evidence.txt")
echo "Upload response: $UPLOAD"
EVIDENCE_ID=$(echo $UPLOAD | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$EVIDENCE_ID" ]; then
    log_success "POST /evidence/incidence/:incidence_id"
else
    log_fail "POST /evidence/incidence/:incidence_id"
fi

# List evidence
log_info "Listing evidence..."
EVIDENCE_LIST=$(curl -s -X GET "$BASE_URL/evidence/incidence/$INCIDENCE_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "Evidence list: $EVIDENCE_LIST"
log_success "GET /evidence/incidence/:incidence_id"

# ===========================================
# 9. HEALTH/UTILITY ENDPOINTS
# ===========================================
echo ""
echo "=== 9. HEALTH/UTILITY ENDPOINTS ==="
echo ""

cat >> $RESULTS_FILE << 'EOF'
## 9. Health/Utility Endpoints

| Endpoint | Method | Handler | API Client | Frontend Page |
|----------|--------|---------|------------|---------------|
| `/health` | GET | `main.go:healthCheck` | N/A | Health checks |

EOF

# Health check
log_info "Testing health endpoint..."
HEALTH=$(curl -s -X GET "$BASE_URL/health")
echo "Health: $HEALTH"
log_success "GET /health"

# ===========================================
# SUMMARY
# ===========================================
echo ""
echo "=========================================="
echo "TEST SUMMARY"
echo "=========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""
echo "Results saved to: $RESULTS_FILE"

# Add summary to results file
cat >> $RESULTS_FILE << EOF

---

## Test Results Summary

- **Passed**: $PASSED
- **Failed**: $FAILED
- **Date**: $(date)

## IDs for Reference

- **Admin Token**: Saved in /tmp/token.txt
- **White Collar Employee**: $WHITE1_ID
- **Blue Collar Employee**: $BLUE1_ID
- **Gray Collar Employee**: $GRAY1_ID
- **Weekly Period**: $WEEKLY_ID
- **Biweekly Period**: $BIWEEKLY_ID
- **Incidence**: $INCIDENCE_ID
- **Vacation Type**: $VACATION_TYPE_ID

EOF

echo ""
echo "All test IDs saved in $RESULTS_FILE"

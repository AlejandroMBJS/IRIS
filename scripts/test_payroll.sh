#!/bin/bash
set -e

BASE_URL="http://localhost:8080/api/v1"

echo "=== 1. Register Admin User ==="
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@iris.com","password":"Test123!@#","full_name":"Admin User","role":"admin","company_name":"IRIS Test Company","company_rfc":"ITT123456ABC"}')
echo "$REGISTER_RESPONSE"

TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to get token"
  exit 1
fi
echo "Token obtained successfully: ${TOKEN:0:50}..."

echo ""
echo "=== 2. Create White Collar Employee ==="
WHITE_RESP=$(curl -s -X POST "$BASE_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number":"WC001",
    "first_name":"Maria","last_name":"Garcia",
    "date_of_birth":"1990-05-15T00:00:00Z","gender":"female",
    "rfc":"GARM900515XXX","curp":"GARM900515MSLRRR09",
    "hire_date":"2024-01-15T00:00:00Z","daily_salary":800.00,
    "employment_status":"active","employee_type":"permanent",
    "collar_type":"white_collar","pay_frequency":"biweekly"
  }')
echo "$WHITE_RESP"
WHITE_ID=$(echo "$WHITE_RESP" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
if [ -z "$WHITE_ID" ]; then
  echo "ERROR: Failed to create white collar employee"
  exit 1
fi
echo "White Collar ID: $WHITE_ID"

echo ""
echo "=== 3. Create Blue Collar Employee ==="
BLUE_RESP=$(curl -s -X POST "$BASE_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number":"BC001",
    "first_name":"Juan","last_name":"Perez",
    "date_of_birth":"1985-03-20T00:00:00Z","gender":"male",
    "rfc":"PERJ850320XXX","curp":"PERJ850320HSLRRN01",
    "hire_date":"2023-06-01T00:00:00Z","daily_salary":400.00,
    "employment_status":"active","employee_type":"permanent",
    "collar_type":"blue_collar","pay_frequency":"weekly","is_sindicalizado":true
  }')
echo "$BLUE_RESP"
BLUE_ID=$(echo "$BLUE_RESP" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Blue Collar ID: $BLUE_ID"

echo ""
echo "=== 4. Create Gray Collar Employee ==="
GRAY_RESP=$(curl -s -X POST "$BASE_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number":"GC001",
    "first_name":"Carlos","last_name":"Lopez",
    "date_of_birth":"1988-09-10T00:00:00Z","gender":"male",
    "rfc":"LOPC880910XXX","curp":"LOPC880910HSLRRS02",
    "hire_date":"2024-02-01T00:00:00Z","daily_salary":500.00,
    "employment_status":"active","employee_type":"permanent",
    "collar_type":"gray_collar","pay_frequency":"weekly","is_sindicalizado":false
  }')
echo "$GRAY_RESP"
GRAY_ID=$(echo "$GRAY_RESP" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Gray Collar ID: $GRAY_ID"

echo ""
echo "=== 5. Create Biweekly Period ==="
PERIOD_RESP=$(curl -s -X POST "$BASE_URL/payroll/periods" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "period_code":"2025-BW24",
    "year":2025,"period_number":24,
    "frequency":"biweekly","period_type":"biweekly",
    "start_date":"2025-12-01T00:00:00Z",
    "end_date":"2025-12-14T00:00:00Z",
    "payment_date":"2025-12-16T00:00:00Z"
  }')
echo "$PERIOD_RESP"
PERIOD_ID=$(echo "$PERIOD_RESP" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Period ID: $PERIOD_ID"

echo ""
echo "=== 6. Insert Prenomina Records Directly ==="
cd /home/amb/iris-payroll-system/backend
sqlite3 iris_payroll.db "INSERT INTO prenomina_metrics (id, employee_id, payroll_period_id, calculation_status, worked_days, regular_hours, overtime_hours, regular_salary, gross_income, created_at, updated_at) VALUES ('$(cat /proc/sys/kernel/random/uuid)', '$WHITE_ID', '$PERIOD_ID', 'approved', 15, 120, 5, 12000, 12500, datetime('now'), datetime('now'));"
sqlite3 iris_payroll.db "INSERT INTO prenomina_metrics (id, employee_id, payroll_period_id, calculation_status, worked_days, regular_hours, overtime_hours, regular_salary, gross_income, created_at, updated_at) VALUES ('$(cat /proc/sys/kernel/random/uuid)', '$BLUE_ID', '$PERIOD_ID', 'approved', 7, 56, 8, 2800, 3200, datetime('now'), datetime('now'));"
sqlite3 iris_payroll.db "INSERT INTO prenomina_metrics (id, employee_id, payroll_period_id, calculation_status, worked_days, regular_hours, overtime_hours, regular_salary, gross_income, created_at, updated_at) VALUES ('$(cat /proc/sys/kernel/random/uuid)', '$GRAY_ID', '$PERIOD_ID', 'approved', 7, 56, 4, 3500, 3700, datetime('now'), datetime('now'));"
echo "Prenomina records inserted"

echo ""
echo "=== 7. Calculate Payroll for All Employees ==="
echo "White Collar:"
curl -s -X POST "$BASE_URL/payroll/calculate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"employee_id\":\"$WHITE_ID\",\"payroll_period_id\":\"$PERIOD_ID\",\"calculate_sdi\":true}"
echo ""
echo "Blue Collar:"
curl -s -X POST "$BASE_URL/payroll/calculate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"employee_id\":\"$BLUE_ID\",\"payroll_period_id\":\"$PERIOD_ID\",\"calculate_sdi\":true}"
echo ""
echo "Gray Collar:"
curl -s -X POST "$BASE_URL/payroll/calculate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"employee_id\":\"$GRAY_ID\",\"payroll_period_id\":\"$PERIOD_ID\",\"calculate_sdi\":true}"
echo ""

echo ""
echo "=== 8. Test PDF Generation for All Collar Types ==="
echo "White Collar PDF:"
curl -s -X GET "$BASE_URL/payroll/payslip/$PERIOD_ID/$WHITE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/payslip_white.pdf
file /tmp/payslip_white.pdf
ls -la /tmp/payslip_white.pdf

echo "Blue Collar PDF:"
curl -s -X GET "$BASE_URL/payroll/payslip/$PERIOD_ID/$BLUE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/payslip_blue.pdf
file /tmp/payslip_blue.pdf
ls -la /tmp/payslip_blue.pdf

echo "Gray Collar PDF:"
curl -s -X GET "$BASE_URL/payroll/payslip/$PERIOD_ID/$GRAY_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/payslip_gray.pdf
file /tmp/payslip_gray.pdf
ls -la /tmp/payslip_gray.pdf

echo ""
echo "=== 9. Test XML (CFDI) Generation for All Collar Types ==="
echo "White Collar XML:"
curl -s -X GET "$BASE_URL/payroll/payslip/$PERIOD_ID/$WHITE_ID?format=xml" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/cfdi_white.xml
file /tmp/cfdi_white.xml
ls -la /tmp/cfdi_white.xml

echo "Blue Collar XML:"
curl -s -X GET "$BASE_URL/payroll/payslip/$PERIOD_ID/$BLUE_ID?format=xml" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/cfdi_blue.xml
file /tmp/cfdi_blue.xml
ls -la /tmp/cfdi_blue.xml

echo "Gray Collar XML:"
curl -s -X GET "$BASE_URL/payroll/payslip/$PERIOD_ID/$GRAY_ID?format=xml" \
  -H "Authorization: Bearer $TOKEN" \
  -o /tmp/cfdi_gray.xml
file /tmp/cfdi_gray.xml
ls -la /tmp/cfdi_gray.xml

echo ""
echo "=== 10. Test Payroll Summary (Period Export) ==="
curl -s -X GET "$BASE_URL/payroll/summary/$PERIOD_ID" \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo ""
echo "=== 11. Test Concept Totals (Financial Report) ==="
curl -s -X GET "$BASE_URL/payroll/concept-totals/$PERIOD_ID" \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo ""
echo "=== Test Complete ==="
echo "PDF files saved to /tmp/payslip_*.pdf"
echo "XML files saved to /tmp/cfdi_*.xml"

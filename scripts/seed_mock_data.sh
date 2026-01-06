#!/bin/bash

# Mock Data Seeder for IRIS Payroll System
# Creates admin user, multiple users with different roles, and employees of all collar types

API_URL="http://localhost:8080/api/v1"

echo "=========================================="
echo "IRIS Payroll System - Mock Data Seeder"
echo "=========================================="

# Step 1: Register Admin User
echo ""
echo "=== Step 1: Registering Admin User ==="
ADMIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "IRIS Tech Solutions SA de CV",
    "company_rfc": "ITS123456789",
    "email": "admin@iris.com",
    "password": "Password123@",
    "full_name": "Administrador Principal",
    "role": "admin"
  }')

echo "$ADMIN_RESPONSE" | jq -r '.user.email // .error // "Error"'
TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.access_token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to get admin token. Trying login..."
  ADMIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email": "admin@iris.com", "password": "Password123@"}')
  TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.access_token')
fi

echo "Admin Token obtained: ${TOKEN:0:50}..."

# Step 2: Create Users with Different Roles
echo ""
echo "=== Step 2: Creating Users with Different Roles ==="

# HR User
echo "Creating HR User..."
curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rh@iris.com",
    "password": "Password123@",
    "full_name": "Maria Garcia (Recursos Humanos)",
    "role": "hr"
  }' | jq -r '.email // .error'

# Accountant User
echo "Creating Accountant User..."
curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "contabilidad@iris.com",
    "password": "Password123@",
    "full_name": "Carlos Lopez (Contabilidad)",
    "role": "accountant"
  }' | jq -r '.email // .error'

# Payroll Staff User
echo "Creating Payroll Staff User..."
curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nominas@iris.com",
    "password": "Password123@",
    "full_name": "Ana Martinez (Nominas)",
    "role": "payroll_staff"
  }' | jq -r '.email // .error'

# Viewer User
echo "Creating Viewer User..."
curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "consulta@iris.com",
    "password": "Password123@",
    "full_name": "Pedro Sanchez (Consulta)",
    "role": "viewer"
  }' | jq -r '.email // .error'

# Step 3: Create Incidence Types
echo ""
echo "=== Step 3: Creating Incidence Types ==="

# Vacation
curl -s -X POST "$API_URL/incidence-types" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Vacaciones",
    "category": "vacation",
    "effect_type": "neutral",
    "is_calculated": false,
    "description": "Dias de vacaciones del empleado"
  }' | jq -r '.name // .error'

# Sick Leave
curl -s -X POST "$API_URL/incidence-types" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Incapacidad por Enfermedad",
    "category": "sick",
    "effect_type": "negative",
    "is_calculated": true,
    "calculation_method": "daily_rate",
    "description": "Incapacidad por enfermedad general"
  }' | jq -r '.name // .error'

# Unexcused Absence
curl -s -X POST "$API_URL/incidence-types" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Falta Injustificada",
    "category": "absence",
    "effect_type": "negative",
    "is_calculated": true,
    "calculation_method": "daily_rate",
    "description": "Falta sin justificacion"
  }' | jq -r '.name // .error'

# Overtime
curl -s -X POST "$API_URL/incidence-types" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tiempo Extra Doble",
    "category": "overtime",
    "effect_type": "positive",
    "is_calculated": true,
    "calculation_method": "hourly_rate",
    "default_value": 2,
    "description": "Horas extras al doble"
  }' | jq -r '.name // .error'

# Delay
curl -s -X POST "$API_URL/incidence-types" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Retardo",
    "category": "delay",
    "effect_type": "negative",
    "is_calculated": false,
    "description": "Llegada tarde al trabajo"
  }' | jq -r '.name // .error'

# Bonus
curl -s -X POST "$API_URL/incidence-types" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bono de Productividad",
    "category": "bonus",
    "effect_type": "positive",
    "is_calculated": false,
    "default_value": 500,
    "description": "Bono por cumplimiento de metas"
  }' | jq -r '.name // .error'

# Step 4: Create White Collar Employees (Biweekly, Monthly)
echo ""
echo "=== Step 4: Creating White Collar Employees ==="

# White Collar 1 - Director
echo "Creating Director..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "WC001",
    "first_name": "Roberto",
    "last_name": "Hernandez",
    "mother_last_name": "Gonzalez",
    "date_of_birth": "1975-03-15T00:00:00Z",
    "gender": "male",
    "rfc": "HEGR750315XXX",
    "curp": "HEGR750315HSLRBT01",
    "nss": "12345678901",
    "hire_date": "2020-01-15T00:00:00Z",
    "daily_salary": 2500.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "white_collar",
    "pay_frequency": "monthly",
    "payment_method": "bank_transfer",
    "bank_name": "BBVA",
    "bank_account": "0123456789",
    "clabe": "012345678901234567"
  }' | jq -r '.first_name + " " + .last_name // .error'

# White Collar 2 - Manager
echo "Creating Manager..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "WC002",
    "first_name": "Laura",
    "last_name": "Ramirez",
    "mother_last_name": "Torres",
    "date_of_birth": "1980-07-22T00:00:00Z",
    "gender": "female",
    "rfc": "RATL800722XXX",
    "curp": "RATL800722MSLMRR02",
    "nss": "23456789012",
    "hire_date": "2021-03-01T00:00:00Z",
    "daily_salary": 1800.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "white_collar",
    "pay_frequency": "biweekly",
    "payment_method": "bank_transfer",
    "bank_name": "Santander",
    "bank_account": "1234567890",
    "clabe": "123456789012345678"
  }' | jq -r '.first_name + " " + .last_name // .error'

# White Collar 3 - Analyst
echo "Creating Analyst..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "WC003",
    "first_name": "Fernando",
    "last_name": "Castillo",
    "mother_last_name": "Mendez",
    "date_of_birth": "1990-11-08T00:00:00Z",
    "gender": "male",
    "rfc": "CAMF901108XXX",
    "curp": "CAMF901108HSLSTR03",
    "nss": "34567890123",
    "hire_date": "2022-06-15T00:00:00Z",
    "daily_salary": 1200.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "white_collar",
    "pay_frequency": "biweekly",
    "payment_method": "bank_transfer",
    "bank_name": "Banorte",
    "bank_account": "2345678901",
    "clabe": "234567890123456789"
  }' | jq -r '.first_name + " " + .last_name // .error'

# White Collar 4 - Administrative
echo "Creating Administrative..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "WC004",
    "first_name": "Patricia",
    "last_name": "Morales",
    "mother_last_name": "Flores",
    "date_of_birth": "1985-04-25T00:00:00Z",
    "gender": "female",
    "rfc": "MOFP850425XXX",
    "curp": "MOFP850425MSLRLT04",
    "nss": "45678901234",
    "hire_date": "2019-09-01T00:00:00Z",
    "daily_salary": 900.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "white_collar",
    "pay_frequency": "biweekly",
    "payment_method": "bank_transfer",
    "bank_name": "HSBC",
    "bank_account": "3456789012",
    "clabe": "345678901234567890"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Step 5: Create Blue Collar Employees (Weekly)
echo ""
echo "=== Step 5: Creating Blue Collar Employees ==="

# Blue Collar 1 - Production Operator
echo "Creating Production Operator 1..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "BC001",
    "first_name": "Miguel",
    "last_name": "Rodriguez",
    "mother_last_name": "Silva",
    "date_of_birth": "1988-06-12T00:00:00Z",
    "gender": "male",
    "rfc": "ROSM880612XXX",
    "curp": "ROSM880612HSLDRG05",
    "nss": "56789012345",
    "hire_date": "2023-01-10T00:00:00Z",
    "daily_salary": 450.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "blue_collar",
    "pay_frequency": "weekly",
    "is_sindicalizado": true,
    "payment_method": "bank_transfer",
    "bank_name": "Bancoppel",
    "bank_account": "4567890123",
    "clabe": "456789012345678901"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Blue Collar 2 - Assembly Worker
echo "Creating Assembly Worker..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "BC002",
    "first_name": "Jose",
    "last_name": "Martinez",
    "mother_last_name": "Luna",
    "date_of_birth": "1992-02-28T00:00:00Z",
    "gender": "male",
    "rfc": "MALJ920228XXX",
    "curp": "MALJ920228HSLRTN06",
    "nss": "67890123456",
    "hire_date": "2023-03-20T00:00:00Z",
    "daily_salary": 420.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "blue_collar",
    "pay_frequency": "weekly",
    "is_sindicalizado": true,
    "payment_method": "bank_transfer",
    "bank_name": "Azteca",
    "bank_account": "5678901234",
    "clabe": "567890123456789012"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Blue Collar 3 - Warehouse Worker
echo "Creating Warehouse Worker..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "BC003",
    "first_name": "Rosa",
    "last_name": "Diaz",
    "mother_last_name": "Ortega",
    "date_of_birth": "1995-09-15T00:00:00Z",
    "gender": "female",
    "rfc": "DIOR950915XXX",
    "curp": "DIOR950915MSLZRS07",
    "nss": "78901234567",
    "hire_date": "2024-01-08T00:00:00Z",
    "daily_salary": 380.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "blue_collar",
    "pay_frequency": "weekly",
    "is_sindicalizado": true,
    "payment_method": "bank_transfer",
    "bank_name": "BBVA",
    "bank_account": "6789012345",
    "clabe": "678901234567890123"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Blue Collar 4 - Machine Operator
echo "Creating Machine Operator..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "BC004",
    "first_name": "Antonio",
    "last_name": "Perez",
    "mother_last_name": "Vargas",
    "date_of_birth": "1987-12-03T00:00:00Z",
    "gender": "male",
    "rfc": "PEVA871203XXX",
    "curp": "PEVA871203HSLRRN08",
    "nss": "89012345678",
    "hire_date": "2022-08-15T00:00:00Z",
    "daily_salary": 500.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "blue_collar",
    "pay_frequency": "weekly",
    "is_sindicalizado": true,
    "payment_method": "bank_transfer",
    "bank_name": "Santander",
    "bank_account": "7890123456",
    "clabe": "789012345678901234"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Blue Collar 5 - General Worker
echo "Creating General Worker..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "BC005",
    "first_name": "Carmen",
    "last_name": "Gutierrez",
    "mother_last_name": "Salazar",
    "date_of_birth": "1998-05-20T00:00:00Z",
    "gender": "female",
    "rfc": "GUSC980520XXX",
    "curp": "GUSC980520MSLTRM09",
    "nss": "90123456789",
    "hire_date": "2024-02-01T00:00:00Z",
    "daily_salary": 350.00,
    "employment_status": "active",
    "employee_type": "temporary",
    "collar_type": "blue_collar",
    "pay_frequency": "weekly",
    "is_sindicalizado": false,
    "payment_method": "cash"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Step 6: Create Gray Collar Employees (Biweekly)
echo ""
echo "=== Step 6: Creating Gray Collar Employees ==="

# Gray Collar 1 - Technician
echo "Creating Technician..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "GC001",
    "first_name": "Eduardo",
    "last_name": "Reyes",
    "mother_last_name": "Nunez",
    "date_of_birth": "1983-08-10T00:00:00Z",
    "gender": "male",
    "rfc": "RENE830810XXX",
    "curp": "RENE830810HSLYSD10",
    "nss": "01234567890",
    "hire_date": "2021-05-01T00:00:00Z",
    "daily_salary": 750.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "gray_collar",
    "pay_frequency": "biweekly",
    "payment_method": "bank_transfer",
    "bank_name": "Citibanamex",
    "bank_account": "8901234567",
    "clabe": "890123456789012345"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Gray Collar 2 - Supervisor
echo "Creating Supervisor..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "GC002",
    "first_name": "Martha",
    "last_name": "Aguilar",
    "mother_last_name": "Rojas",
    "date_of_birth": "1979-01-25T00:00:00Z",
    "gender": "female",
    "rfc": "AURM790125XXX",
    "curp": "AURM790125MSLGRT11",
    "nss": "11234567890",
    "hire_date": "2018-11-15T00:00:00Z",
    "daily_salary": 950.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "gray_collar",
    "pay_frequency": "biweekly",
    "payment_method": "bank_transfer",
    "bank_name": "Banorte",
    "bank_account": "9012345678",
    "clabe": "901234567890123456"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Gray Collar 3 - Quality Inspector
echo "Creating Quality Inspector..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "GC003",
    "first_name": "Oscar",
    "last_name": "Campos",
    "mother_last_name": "Rivera",
    "date_of_birth": "1991-04-18T00:00:00Z",
    "gender": "male",
    "rfc": "CARO910418XXX",
    "curp": "CARO910418HSLMPS12",
    "nss": "22345678901",
    "hire_date": "2022-02-28T00:00:00Z",
    "daily_salary": 680.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "gray_collar",
    "pay_frequency": "biweekly",
    "payment_method": "bank_transfer",
    "bank_name": "BBVA",
    "bank_account": "0123456780",
    "clabe": "012345678012345678"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Gray Collar 4 - Maintenance Tech
echo "Creating Maintenance Technician..."
curl -s -X POST "$API_URL/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "GC004",
    "first_name": "Diana",
    "last_name": "Vega",
    "mother_last_name": "Paredes",
    "date_of_birth": "1989-10-30T00:00:00Z",
    "gender": "female",
    "rfc": "VEPD891030XXX",
    "curp": "VEPD891030MSLGRN13",
    "nss": "33456789012",
    "hire_date": "2020-07-20T00:00:00Z",
    "daily_salary": 720.00,
    "employment_status": "active",
    "employee_type": "permanent",
    "collar_type": "gray_collar",
    "pay_frequency": "biweekly",
    "payment_method": "bank_transfer",
    "bank_name": "Scotiabank",
    "bank_account": "1234567891",
    "clabe": "123456789123456789"
  }' | jq -r '.first_name + " " + .last_name // .error'

# Step 7: Create Payroll Periods
echo ""
echo "=== Step 7: Creating Payroll Periods ==="

# Weekly period for blue collar
echo "Creating Weekly Period..."
curl -s -X POST "$API_URL/payroll/periods" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "period_code": "SEM-2025-50",
    "year": 2025,
    "period_number": 50,
    "frequency": "weekly",
    "period_type": "weekly",
    "start_date": "2025-12-02",
    "end_date": "2025-12-08",
    "payment_date": "2025-12-10",
    "description": "Semana 50 - Diciembre 2025"
  }' | jq -r '.period_code // .error'

# Biweekly period for white/gray collar
echo "Creating Biweekly Period..."
curl -s -X POST "$API_URL/payroll/periods" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "period_code": "QNA-2025-24",
    "year": 2025,
    "period_number": 24,
    "frequency": "biweekly",
    "period_type": "biweekly",
    "start_date": "2025-12-01",
    "end_date": "2025-12-15",
    "payment_date": "2025-12-17",
    "description": "Quincena 24 - Primera de Diciembre 2025"
  }' | jq -r '.period_code // .error'

# Step 8: Verify data
echo ""
echo "=== Step 8: Verifying Data ==="

echo ""
echo "Users created:"
curl -s -X GET "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "  - " + .email + " (" + .role + ")"'

echo ""
echo "Employees by collar type:"
EMPLOYEES=$(curl -s -X GET "$API_URL/employees" -H "Authorization: Bearer $TOKEN")
echo "$EMPLOYEES" | jq -r '.employees[] | "  - [" + .collar_type + "] " + .employee_number + ": " + .first_name + " " + .last_name + " - $" + (.daily_salary|tostring) + "/day"'

echo ""
echo "Incidence Types:"
curl -s -X GET "$API_URL/incidence-types" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "  - " + .name + " (" + .category + ", " + .effect_type + ")"'

echo ""
echo "Payroll Periods:"
curl -s -X GET "$API_URL/payroll/periods" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "  - " + .period_code + " (" + .period_type + ") " + .start_date + " to " + .end_date'

echo ""
echo "=========================================="
echo "Mock Data Seeding Complete!"
echo "=========================================="
echo ""
echo "Test Accounts:"
echo "  Admin:    admin@iris.com / Password123@"
echo "  HR:       rh@iris.com / Password123@"
echo "  Account:  contabilidad@iris.com / Password123@"
echo "  Payroll:  nominas@iris.com / Password123@"
echo "  Viewer:   consulta@iris.com / Password123@"
echo ""
echo "Employee Summary:"
echo "  - 4 White Collar (biweekly/monthly)"
echo "  - 5 Blue Collar (weekly)"
echo "  - 4 Gray Collar (biweekly)"
echo ""

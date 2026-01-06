# IRIS Payroll System - IT Operations Guide

## Table of Contents
1. [System Overview](#system-overview)
2. [System Requirements](#system-requirements)
3. [Installation Methods](#installation-methods)
4. [Initial Setup](#initial-setup)
5. [Employee Data Seeding](#employee-data-seeding)
6. [User Management](#user-management)
7. [Excel/XLS Employee Upload](#excelxls-employee-upload)
8. [Backup & Recovery](#backup--recovery)
9. [Monitoring & Maintenance](#monitoring--maintenance)
10. [Troubleshooting](#troubleshooting)

---

## System Overview

IRIS Payroll is a comprehensive Mexican payroll management system that handles:
- Employee management (White/Blue/Gray collar workers)
- Payroll calculations (ISR, IMSS, INFONAVIT deductions)
- Vacation tracking per Ley Federal del Trabajo (LFT)
- Incidence management (absences, overtime, bonuses)
- PDF payslip generation
- CFDI XML generation for SAT compliance

### Architecture
```
Frontend (Next.js) --> Backend API (Go/Gin) --> SQLite Database
        :3000                  :8080              iris_payroll.db
```

---

## System Requirements

### Minimum Requirements
- **OS**: Linux (Ubuntu 20.04+), macOS, Windows 10+
- **CPU**: 2 cores
- **RAM**: 4 GB
- **Disk**: 10 GB free space
- **Network**: Internet access for initial setup

### Software Requirements
- Go 1.21+ (for local development)
- Node.js 18+ (for frontend)
- Docker & Docker Compose (for containerized deployment)
- SQLite3 (included with Go)

---

## Installation Methods

### Method 1: Docker Compose (Recommended for Production)

1. **Clone the repository**
```bash
git clone <repository-url>
cd iris-payroll-system
```

2. **Start services**
```bash
docker-compose up -d
```

3. **Verify services are running**
```bash
docker-compose ps
curl http://localhost:8080/api/v1/health
```

### Method 2: Local Development

#### Backend Setup
```bash
cd backend

# Install Go dependencies
go mod download

# Build the application
go build -o main ./cmd/api

# Run the server
./main
```

#### Frontend Setup
```bash
cd frontend

# Install Node dependencies
pnpm install
# or: npm install

# Run development server
pnpm dev
# or: npm run dev
```

### Method 3: Production Build

#### Backend
```bash
cd backend
CGO_ENABLED=1 go build -ldflags="-s -w" -o iris-payroll ./cmd/api
```

#### Frontend
```bash
cd frontend
pnpm build
pnpm start
```

---

## Initial Setup

### Step 1: Register Admin User and Company

The first user to register creates the company and becomes the admin:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Mi Empresa S.A. de C.V.",
    "company_rfc": "MEM123456ABC",
    "email": "admin@miempresa.com",
    "password": "SecurePass123@",
    "full_name": "Administrador Principal"
  }'
```

**Password Requirements:**
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character (!@#$%^&*())

### Step 2: Login and Get Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@miempresa.com",
    "password": "SecurePass123@"
  }'
```

Save the `access_token` from the response for subsequent API calls.

### Step 3: Configure System Settings (Optional)

```bash
TOKEN="your_access_token_here"

# Update company settings
curl -X PUT http://localhost:8080/api/v1/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "default_work_hours": 8,
    "default_work_days": 6,
    "uma_value": 108.57,
    "minimum_wage": 248.93
  }'
```

---

## Employee Data Seeding

### Required Employee Data Fields

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| employee_number | Yes | Unique identifier | "EMP001" |
| first_name | Yes | First name | "Juan" |
| last_name | Yes | Last name (paterno) | "Perez" |
| mother_last_name | No | Last name (materno) | "Lopez" |
| date_of_birth | Yes | Birth date (ISO 8601) | "1990-05-15T00:00:00Z" |
| gender | Yes | male/female/other | "male" |
| rfc | Yes | RFC (12-13 chars) | "PERJ900515XXX" |
| curp | Yes | CURP (18 chars) | "PERJ900515HSLRLN09" |
| nss | No | IMSS number (11 digits) | "12345678901" |
| hire_date | Yes | Start date | "2024-01-15T00:00:00Z" |
| daily_salary | Yes | Daily wage in MXN | 500.00 |
| collar_type | Yes | Worker type | "white_collar" / "blue_collar" / "gray_collar" |
| pay_frequency | Yes | Payment frequency | "weekly" / "biweekly" / "monthly" |
| employment_status | Yes | Current status | "active" / "inactive" / "on_leave" / "terminated" |
| employee_type | Yes | Contract type | "permanent" / "temporary" / "contractor" / "intern" |

### Collar Types Explained

| Type | Description | Payment Frequency | Union Status |
|------|-------------|-------------------|--------------|
| white_collar | Administrative/Office | Biweekly | Non-unionized |
| blue_collar | Production/Factory | Weekly | Unionized |
| gray_collar | Production/Factory | Weekly | Non-unionized |

### Example: Create Single Employee

```bash
TOKEN="your_access_token_here"

curl -X POST http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "EMP001",
    "first_name": "Juan",
    "last_name": "Perez",
    "mother_last_name": "Lopez",
    "date_of_birth": "1990-05-15T00:00:00Z",
    "gender": "male",
    "rfc": "PELJ900515XXX",
    "curp": "PELJ900515HSLRLN09",
    "nss": "12345678901",
    "hire_date": "2024-01-15T00:00:00Z",
    "daily_salary": 500.00,
    "collar_type": "white_collar",
    "pay_frequency": "biweekly",
    "employment_status": "active",
    "employee_type": "permanent",
    "is_sindicalizado": false,
    "payment_method": "bank_transfer",
    "bank_name": "BBVA",
    "bank_account": "1234567890",
    "clabe": "012180001234567890"
  }'
```

### Batch Employee Seed Script

Create a file `seed_employees.sh`:

```bash
#!/bin/bash

# Configuration
API_URL="http://localhost:8080/api/v1"
EMAIL="admin@miempresa.com"
PASSWORD="SecurePass123@"

# Login and get token
TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  | jq -r '.access_token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Error: Failed to login"
  exit 1
fi

echo "Login successful, token obtained"

# Function to create employee
create_employee() {
  local data=$1
  response=$(curl -s -X POST "$API_URL/employees" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$data")
  echo "$response" | jq -r '.employee_number // .error'
}

# Create employees
echo "Creating White Collar Employee..."
create_employee '{
  "employee_number": "ADM001",
  "first_name": "Maria",
  "last_name": "Garcia",
  "date_of_birth": "1985-03-20T00:00:00Z",
  "gender": "female",
  "rfc": "GAMA850320XXX",
  "curp": "GAMA850320MSLRRA09",
  "hire_date": "2023-06-01T00:00:00Z",
  "daily_salary": 800.00,
  "collar_type": "white_collar",
  "pay_frequency": "biweekly",
  "employment_status": "active",
  "employee_type": "permanent"
}'

echo "Creating Blue Collar Employee..."
create_employee '{
  "employee_number": "OPE001",
  "first_name": "Pedro",
  "last_name": "Martinez",
  "date_of_birth": "1988-07-10T00:00:00Z",
  "gender": "male",
  "rfc": "MAMP880710XXX",
  "curp": "MAMP880710HSLRTR09",
  "hire_date": "2022-01-15T00:00:00Z",
  "daily_salary": 350.00,
  "collar_type": "blue_collar",
  "pay_frequency": "weekly",
  "employment_status": "active",
  "employee_type": "permanent",
  "is_sindicalizado": true
}'

echo "Creating Gray Collar Employee..."
create_employee '{
  "employee_number": "TEC001",
  "first_name": "Ana",
  "last_name": "Rodriguez",
  "date_of_birth": "1992-11-25T00:00:00Z",
  "gender": "female",
  "rfc": "ROAA921125XXX",
  "curp": "ROAA921125MSLDRN09",
  "hire_date": "2024-03-01T00:00:00Z",
  "daily_salary": 450.00,
  "collar_type": "gray_collar",
  "pay_frequency": "weekly",
  "employment_status": "active",
  "employee_type": "permanent",
  "is_sindicalizado": false
}'

echo "Employee seeding completed!"
```

Run with:
```bash
chmod +x seed_employees.sh
./seed_employees.sh
```

---

## User Management

### User Roles

| Role | Permissions |
|------|-------------|
| admin | Full system access, user management, settings |
| manager | Payroll processing, employee management, reports |
| operator | Basic data entry, view reports |
| viewer | Read-only access |

### Creating Additional Users

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "payroll@miempresa.com",
    "password": "PayrollUser123@",
    "full_name": "Operador de Nomina",
    "role": "operator",
    "employee_id": "uuid-of-employee-if-applicable"
  }'
```

### Linking Users to Employees

When a payroll team member is also an employee:

1. First create the employee record
2. Then create or update the user with the `employee_id` field

```bash
# Get employee ID
EMPLOYEE_ID=$(curl -s http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer $TOKEN" | jq -r '.employees[] | select(.employee_number=="ADM001") | .id')

# Link user to employee
curl -X PUT http://localhost:8080/api/v1/users/{user_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"employee_id\": \"$EMPLOYEE_ID\"}"
```

---

## Excel/XLS Employee Upload

### Template Format

Download the employee template or create an Excel file with these columns:

| Column | Header Name | Required | Format |
|--------|------------|----------|--------|
| A | employee_number | Yes | Text (e.g., "EMP001") |
| B | first_name | Yes | Text |
| C | last_name | Yes | Text |
| D | mother_last_name | No | Text |
| E | date_of_birth | Yes | Date (YYYY-MM-DD) |
| F | gender | Yes | male/female/other |
| G | rfc | Yes | 12-13 characters |
| H | curp | Yes | 18 characters |
| I | nss | No | 11 digits |
| J | hire_date | Yes | Date (YYYY-MM-DD) |
| K | daily_salary | Yes | Number (e.g., 500.00) |
| L | collar_type | Yes | white_collar/blue_collar/gray_collar |
| M | pay_frequency | Yes | weekly/biweekly/monthly |
| N | employment_status | Yes | active/inactive |
| O | employee_type | Yes | permanent/temporary/contractor/intern |
| P | is_sindicalizado | No | true/false |
| Q | bank_name | No | Text |
| R | bank_account | No | Text |
| S | clabe | No | 18 digits |

### Upload via API

```bash
curl -X POST http://localhost:8080/api/v1/employees/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@employees.xlsx"
```

### Upload via Frontend

1. Navigate to **Employees > Import**
2. Download template if needed
3. Fill in employee data
4. Upload the file
5. Review validation results
6. Confirm import

### Validation Rules

Before import, the system validates:
- RFC format: `^[A-Z&N]{3,4}[0-9]{6}[A-Z0-9]{3}$`
- CURP format: `^[A-Z]{4}[0-9]{6}[HM][A-Z]{5}[A-Z0-9]{2}$`
- NSS format: `^[0-9]{11}$`
- No duplicate employee_number, RFC, or CURP
- Valid dates (birth date before hire date)
- Positive salary values

---

## Backup & Recovery

### Database Backup

```bash
# Create backup
cp backend/iris_payroll.db backups/iris_payroll_$(date +%Y%m%d_%H%M%S).db

# Automated daily backup (add to crontab)
0 2 * * * cp /path/to/iris_payroll.db /path/to/backups/iris_payroll_$(date +\%Y\%m\%d).db
```

### Restore from Backup

```bash
# Stop the application
docker-compose down
# or: pkill -f iris-payroll

# Restore database
cp backups/iris_payroll_20240115.db backend/iris_payroll.db

# Restart application
docker-compose up -d
# or: ./main
```

### Export Data

```bash
# Export employees to JSON
curl -s http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer $TOKEN" > employees_backup.json

# Export payroll data
curl -s http://localhost:8080/api/v1/payroll/periods \
  -H "Authorization: Bearer $TOKEN" > periods_backup.json
```

---

## Monitoring & Maintenance

### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "version": "1.0.0"
}
```

### Log Monitoring

```bash
# Docker logs
docker-compose logs -f backend

# Local logs
tail -f backend/logs/app.log
```

### Performance Monitoring

Monitor these endpoints:
- `/api/v1/health` - System health
- `/api/v1/employees/stats` - Employee statistics
- Database file size: `ls -lh backend/iris_payroll.db`

### Routine Maintenance Tasks

| Task | Frequency | Command |
|------|-----------|---------|
| Database backup | Daily | `cp iris_payroll.db backups/` |
| Clear old logs | Weekly | `find logs/ -mtime +30 -delete` |
| Vacuum database | Monthly | `sqlite3 iris_payroll.db "VACUUM;"` |
| Check disk space | Weekly | `df -h` |

---

## Troubleshooting

### Common Issues

#### 1. Database Locked
```
Error: database is locked
```
**Solution**: Only one process should access SQLite. Stop duplicate processes:
```bash
lsof -i:8080
pkill -f "main"
```

#### 2. Port Already in Use
```
Error: listen tcp :8080: bind: address already in use
```
**Solution**:
```bash
lsof -ti:8080 | xargs kill -9
```

#### 3. Invalid Token
```
Error: token is expired
```
**Solution**: Login again to get a new token. Tokens expire after 24 hours.

#### 4. RFC/CURP Validation Failed
**Solution**: Ensure format matches:
- RFC: 12-13 alphanumeric characters
- CURP: 18 characters with specific format

#### 5. Frontend Can't Connect to Backend
**Solution**: Check CORS settings and API URL:
```bash
# Verify backend is running
curl http://localhost:8080/api/v1/health

# Check frontend .env
cat frontend/.env.local
# Should have: NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

### Contact Support

For additional support:
- Check logs at `backend/logs/`
- Review API documentation at `http://localhost:8080/swagger/index.html`
- Contact: soporte@miempresa.com

---

## Appendix

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | API server port |
| DATABASE_PATH | ./iris_payroll.db | SQLite database path |
| JWT_SECRET | (generated) | JWT signing key |
| LOG_LEVEL | info | Logging level |

### API Quick Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| /auth/register | POST | Register company/admin |
| /auth/login | POST | Login |
| /employees | GET | List employees |
| /employees | POST | Create employee |
| /employees/import | POST | Bulk import from Excel |
| /payroll/periods | GET | List pay periods |
| /payroll/calculate | POST | Calculate payroll |
| /payroll/payslip/{period}/{employee} | GET | Get payslip |

---

*Document Version: 1.0*
*Last Updated: December 2024*

# IRIS Payroll System

Sistema de Administracion de Nomina Mexicana - Mexican Payroll Management System

A complete payroll management system designed specifically for Mexican businesses, featuring comprehensive payroll calculations compliant with Mexican labor laws (LFT), IMSS, SAT, and INFONAVIT regulations. Configured for San Luis Potosi, Mexico.

---

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Architecture](#architecture)
4. [Technology Stack](#technology-stack)
5. [Project Structure](#project-structure)
6. [Prerequisites](#prerequisites)
7. [Installation](#installation)
8. [Configuration](#configuration)
9. [Running the Application](#running-the-application)
10. [API Documentation](#api-documentation)
11. [Database Models](#database-models)
12. [Payroll Calculations](#payroll-calculations)
13. [Mexican Payroll Compliance](#mexican-payroll-compliance)
14. [Employee Types (Collar Types)](#employee-types-collar-types)
15. [Reports and Exports](#reports-and-exports)
16. [Frontend Pages](#frontend-pages)
17. [Development Guide](#development-guide)
18. [Troubleshooting](#troubleshooting)
19. [License](#license)

---

## Overview

IRIS Payroll System is a comprehensive payroll management solution designed for Mexican companies operating under Mexican labor laws. The system handles:

- Employee management with Mexican ID validation (RFC, CURP, NSS)
- Payroll calculations with ISR (Income Tax), IMSS (Social Security), and INFONAVIT deductions
- CFDI 4.0 compliant XML generation with Nomina 1.2 complement
- PDF payslip generation
- Multiple payment frequencies (weekly, biweekly, monthly)
- Employee classification by collar type (white, blue, gray)
- Regional configuration for San Luis Potosi

---

## Features

### Backend Features

- **Authentication & Authorization**: JWT-based authentication with role-based access control (admin, hr, employee)
- **Employee Management**: Full CRUD operations with Mexican ID validation
- **Payroll Processing**:
  - Individual and bulk payroll calculations
  - Support for multiple payment frequencies
  - Automatic tax calculations (ISR with employment subsidy)
  - IMSS employee and employer contributions
  - INFONAVIT deductions
  - Overtime calculations (double and triple time)
- **Pre-payroll (Prenomina)**: Work metrics tracking before payroll processing
- **CFDI Generation**: XML compliant with SAT CFDI 4.0 and Nomina 1.2
- **PDF Payslips**: Professional payslip generation in PDF format
- **Reports**: Payroll summaries, concept totals, and employee reports
- **Configuration**: JSON-based modular configuration for easy updates

### Frontend Features

- **Dashboard**: Overview of payroll status and key metrics
- **Employee Management**: Create, edit, view, and manage employees
- **Payroll Processing**: Calculate and process payroll with visual feedback
- **Reports**: Generate and view payroll reports
- **Configuration**: System settings and payroll configuration
- **Dark Mode**: Full dark mode support with next-themes
- **Responsive Design**: Works on desktop and mobile devices

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend                              │
│                   (Next.js 14 + TypeScript)                 │
│                         :3000                                │
└─────────────────────────┬───────────────────────────────────┘
                          │ HTTP/REST
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                        Backend                               │
│                    (Go + Gin + GORM)                        │
│                         :8080                                │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │   API    │  │ Services │  │   Repos  │  │  Models  │    │
│  │ Handlers │──│  Layer   │──│  Layer   │──│  (GORM)  │    │
│  └──────────┘  └──────────┘  └──────────┘  └────┬─────┘    │
└─────────────────────────────────────────────────┼───────────┘
                                                  │
                          ┌───────────────────────┴──────────┐
                          │         SQLite Database          │
                          │        (iris_payroll.db)         │
                          └──────────────────────────────────┘
```

---

## Technology Stack

### Backend
| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.18+ | Programming language |
| Gin | 1.9+ | Web framework |
| GORM | 1.25+ | ORM for database operations |
| SQLite | 3.x | Database (default) |
| PostgreSQL | 12+ | Database (production option) |
| JWT | - | Authentication |
| gofpdf | - | PDF generation |

### Frontend
| Technology | Version | Purpose |
|------------|---------|---------|
| Next.js | 14.x | React framework with App Router |
| TypeScript | 5.x | Type-safe JavaScript |
| Tailwind CSS | 4.x | Utility-first CSS |
| Radix UI | - | Accessible UI components |
| React Hook Form | - | Form management |
| Zod | - | Schema validation |
| Lucide React | - | Icons |
| next-themes | - | Dark mode support |

---

## Project Structure

```
iris-payroll-system/
├── backend/                          # Go backend application
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Application entry point
│   ├── configs/                      # Configuration files
│   │   ├── payroll/                 # Payroll configuration
│   │   │   ├── main.json            # Master config file
│   │   │   ├── official_values.json # UMA, minimum wages
│   │   │   ├── regional_slp.json    # San Luis Potosi settings
│   │   │   ├── contribution_rates.json # IMSS/INFONAVIT rates
│   │   │   ├── labor_concepts.json  # Labor law concepts
│   │   │   └── calculation_tables.json # Reference tables
│   │   ├── tables/                  # Calculation tables
│   │   │   ├── isr_monthly_2025.json
│   │   │   ├── isr_biweekly_2025.json
│   │   │   ├── subsidy_2025.json
│   │   │   └── vacations_2025.json
│   │   └── holidays/
│   │       └── slp_2025.json        # Holiday calendar
│   ├── internal/                    # Private application code
│   │   ├── api/                     # HTTP handlers
│   │   │   ├── router.go            # Route definitions
│   │   │   ├── auth_handler.go      # Authentication endpoints
│   │   │   ├── employee_handler.go  # Employee endpoints
│   │   │   ├── payroll_handler.go   # Payroll endpoints
│   │   │   ├── payroll_period_handler.go
│   │   │   ├── report_handler.go    # Report endpoints
│   │   │   └── catalog_handler.go   # Catalog endpoints
│   │   ├── config/                  # Configuration loading
│   │   ├── database/                # Database connection
│   │   ├── dtos/                    # Data Transfer Objects
│   │   ├── middleware/              # HTTP middleware
│   │   ├── models/                  # Database models
│   │   │   ├── base.go
│   │   │   ├── user.go
│   │   │   ├── company.go
│   │   │   ├── employee.go
│   │   │   ├── payroll.go
│   │   │   ├── payroll_period.go
│   │   │   └── cfdi.go
│   │   ├── repositories/            # Data access layer
│   │   ├── services/                # Business logic
│   │   │   ├── auth_service.go
│   │   │   ├── employee_service.go
│   │   │   ├── payroll_service.go
│   │   │   ├── prenomina_service.go
│   │   │   ├── tax_calculation_service.go
│   │   │   ├── cfdi_service.go
│   │   │   └── report_service.go
│   │   └── utils/                   # Utility functions
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── .env.example
│
├── frontend/                        # Next.js frontend application
│   ├── app/                         # App Router pages
│   │   ├── layout.tsx               # Root layout
│   │   ├── page.tsx                 # Home page
│   │   ├── globals.css              # Global styles
│   │   ├── auth/                    # Authentication pages
│   │   │   └── login/
│   │   ├── dashboard/               # Dashboard page
│   │   ├── employees/               # Employee management
│   │   ├── payroll/                 # Payroll processing
│   │   └── configuration/           # System settings
│   ├── components/                  # React components
│   │   ├── ui/                      # UI primitives
│   │   └── ...                      # Feature components
│   ├── hooks/                       # Custom React hooks
│   ├── lib/                         # Utilities and API client
│   ├── next.config.js
│   ├── package.json
│   └── tsconfig.json
│
├── scripts/                         # Utility scripts
│   ├── setup.sh                     # Initial setup
│   ├── start.sh                     # Start all services
│   └── stop.sh                      # Stop all services
│
├── README.md                        # This file
├── QUICKSTART.md                    # Quick start guide
└── docker-compose.yml               # Docker configuration
```

---

## Prerequisites

### Backend Requirements
- Go 1.18 or higher
- SQLite 3.x (comes with most systems) or PostgreSQL 12+

### Frontend Requirements
- Node.js 18+ or 20+
- pnpm (recommended) or npm

### Development Tools (Optional)
- Git
- Docker & Docker Compose
- VS Code with Go and TypeScript extensions

---

## Installation

### Quick Setup (Recommended)

```bash
# Clone the repository
cd iris-payroll-system

# Run the setup script
chmod +x scripts/setup.sh
./scripts/setup.sh
```

### Manual Installation

#### Backend Setup

```bash
# Navigate to backend directory
cd backend

# Download Go dependencies
go mod download

# Build the application
go build -o iris-payroll ./cmd/server

# Copy and configure environment
cp .env.example .env
# Edit .env with your settings
```

#### Frontend Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
pnpm install

# Copy and configure environment
cp .env.local.example .env.local
# Edit .env.local with your settings
```

---

## Configuration

### Backend Environment Variables (.env)

```env
# Application
APP_ENV=development          # development, staging, production
APP_PORT=8080               # Server port
DEBUG=true                  # Enable debug logging

# Database
DB_TYPE=sqlite              # sqlite or postgres
DB_NAME=iris_payroll.db     # Database file name (SQLite)
# DB_HOST=localhost         # PostgreSQL host
# DB_PORT=5432              # PostgreSQL port
# DB_USER=iris              # PostgreSQL user
# DB_PASSWORD=password      # PostgreSQL password
# DB_SSLMODE=disable        # PostgreSQL SSL mode

# JWT Authentication
JWT_SECRET=your-super-secret-key-change-in-production
JWT_ACCESS_TOKEN_HOURS=24
JWT_REFRESH_TOKEN_HOURS=168

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

### Frontend Environment Variables (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

### Payroll Configuration Files

The payroll system uses JSON configuration files in `backend/configs/`:

| File | Description |
|------|-------------|
| `payroll/main.json` | Master configuration referencing all other files |
| `payroll/official_values.json` | UMA, minimum wages, fiscal year settings |
| `payroll/regional_slp.json` | San Luis Potosi specific settings |
| `payroll/contribution_rates.json` | IMSS and INFONAVIT contribution rates |
| `payroll/labor_concepts.json` | Labor law concepts (aguinaldo, vacations, etc.) |
| `tables/isr_monthly_2025.json` | Monthly ISR withholding table |
| `tables/isr_biweekly_2025.json` | Biweekly ISR withholding table |
| `tables/subsidy_2025.json` | Employment subsidy table |
| `tables/vacations_2025.json` | Vacation days by years of service |

---

## Running the Application

### Using Scripts (Recommended)

```bash
# Start all services
./scripts/start.sh

# Stop all services
./scripts/stop.sh
```

### Manual Start

#### Start Backend

```bash
cd backend
./iris-payroll
# Or during development:
go run cmd/server/main.go
```

The backend will start on `http://localhost:8080`

#### Start Frontend

```bash
cd frontend
pnpm dev
```

The frontend will start on `http://localhost:3000`

### First-Time Setup

1. Start the backend (database auto-migrates on first run)
2. Register a company and admin user:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Mi Empresa SA de CV",
    "company_rfc": "MEM123456789",
    "email": "admin@miempresa.com",
    "password": "Password123!",
    "full_name": "Administrador",
    "role": "admin"
  }'
```

3. Login via frontend at `http://localhost:3000/auth/login`

---

## API Documentation

### Base URL

```
http://localhost:8080/api/v1
```

### Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <access_token>
```

### Role-Based Access Control

The system supports 5 user roles with different access levels:

| Role | Description | Access Level |
|------|-------------|--------------|
| `admin` | Full system access | All endpoints including user management |
| `hr` | Human Resources | Employees, incidences, reports |
| `accountant` | Accounting | Payroll, reports, catalogs |
| `payroll_staff` | Payroll Processing | Payroll calculations, periods |
| `viewer` | Read-Only | View all data, no modifications |

### Endpoints

#### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register company and admin user |
| POST | `/auth/login` | Login and get tokens |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password with token |

#### Authentication (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/logout` | Logout user |
| POST | `/auth/change-password` | Change user password |
| GET | `/auth/profile` | Get current user profile |
| PUT | `/auth/profile` | Update user profile |

#### User Management (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users` | List all users |
| POST | `/users` | Create new user |
| DELETE | `/users/:id` | Delete user |
| PATCH | `/users/:id/toggle-active` | Activate/deactivate user |

#### Employees

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/employees` | List employees (paginated) |
| GET | `/employees/:id` | Get employee by ID |
| POST | `/employees` | Create new employee |
| PUT | `/employees/:id` | Update employee |
| DELETE | `/employees/:id` | Soft delete employee |
| POST | `/employees/:id/terminate` | Terminate employee |
| PUT | `/employees/:id/salary` | Update employee salary |
| GET | `/employees/stats` | Get employee statistics |
| POST | `/employees/validate-ids` | Validate RFC, CURP, NSS |
| GET | `/employees/:id/incidences` | Get employee incidences |
| GET | `/employees/:id/vacation-balance` | Get vacation balance |
| GET | `/employees/:id/absence-summary` | Get absence summary |

#### Incidence Types (Catalogs)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/incidence-types` | List all incidence types |
| POST | `/incidence-types` | Create incidence type |
| PUT | `/incidence-types/:id` | Update incidence type |
| DELETE | `/incidence-types/:id` | Delete incidence type |

**Incidence Categories:**
- `absence` - Unexcused absence
- `sick` - Sick leave
- `vacation` - Vacation days
- `overtime` - Extra hours worked
- `delay` - Late arrival
- `bonus` - Additional bonus
- `deduction` - Salary deduction
- `other` - Other incidences

**Effect Types:**
- `positive` - Adds to income (bonuses, overtime)
- `negative` - Deducts from income (absences, deductions)
- `neutral` - No payroll effect (informational)

#### Incidences

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/incidences` | List all incidences (with filters) |
| GET | `/incidences/:id` | Get incidence by ID |
| POST | `/incidences` | Create new incidence |
| PUT | `/incidences/:id` | Update incidence |
| DELETE | `/incidences/:id` | Delete incidence |
| POST | `/incidences/:id/approve` | Approve incidence |
| POST | `/incidences/:id/reject` | Reject incidence |

**Incidence Status Flow:**
`pending` -> `approved` -> `processed`
`pending` -> `rejected`

#### Payroll Periods

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/payroll/periods` | List payroll periods |
| GET | `/payroll/periods/:id` | Get period by ID |
| POST | `/payroll/periods` | Create payroll period |

#### Payroll Processing

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/payroll/calculate` | Calculate individual payroll |
| POST | `/payroll/bulk-calculate` | Calculate payroll for multiple employees |
| GET | `/payroll/calculation/:period_id/:employee_id` | Get calculation details |
| GET | `/payroll/period/:period_id` | Get all calculations for period |
| GET | `/payroll/payslip/:periodId/:employeeId` | Generate payslip (PDF/XML) |
| GET | `/payroll/summary/:periodId` | Get payroll summary |
| POST | `/payroll/approve/:periodId` | Approve payroll period |
| POST | `/payroll/payment/:periodId` | Process payment |
| GET | `/payroll/payment/:period_id` | Get payment status |
| GET | `/payroll/concept-totals/:periodId` | Get concept totals |

**Payslip Formats:**
- `?format=pdf` - PDF payslip (Recibo de Nomina)
- `?format=xml` - CFDI 4.0 with Nomina 1.2 complement

#### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/reports/generate` | Generate custom report |
| GET | `/reports/history` | Get report generation history |

#### Catalogs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/catalogs/concepts` | List payroll concepts |
| POST | `/catalogs/concepts` | Create payroll concept |
| GET | `/catalogs/incidence-types` | List incidence types |

---

## Database Models

### Core Models

#### User
- Authentication and authorization
- Roles: admin, hr, employee
- Company association

#### Company
- Business information
- RFC, legal name, address
- IMSS employer registration

#### Employee
- Personal information (name, DOB, gender)
- Mexican IDs (RFC, CURP, NSS)
- Employment details (hire date, status, type)
- Salary information
- Collar type classification

#### PayrollPeriod
- Period definition (start, end, payment date)
- Frequency (weekly, biweekly, monthly)
- Status (draft, open, calculating, closed, paid)

#### PayrollCalculation
- Calculated payroll for an employee in a period
- Incomes (salary, overtime, bonuses)
- Deductions (ISR, IMSS, INFONAVIT, loans)
- Benefits (food vouchers, savings fund)
- Totals (gross, net, employer contributions)

#### PayrollDetail
- Line items for each payroll calculation
- Concept, type, amount, SAT code

#### EmployerContribution
- Employer's IMSS contributions
- INFONAVIT employer contribution
- State payroll tax

---

## Payroll Calculations

### Calculation Flow

1. **Pre-payroll (Prenomina)**
   - Collect work metrics (worked days, hours, overtime)
   - Record absences, delays, leaves

2. **Payroll Calculation**
   - Calculate gross income (salary + overtime + bonuses)
   - Calculate IMSS employee contributions
   - Calculate ISR withholding (with employment subsidy)
   - Calculate INFONAVIT deductions
   - Calculate net pay

3. **Approval & Payment**
   - Review and approve calculations
   - Process payment
   - Generate payslips and CFDI

### Income Calculations

```
Regular Salary = Daily Salary × Worked Days
Overtime (Double) = Hourly Rate × 2 × Double Overtime Hours
Overtime (Triple) = Hourly Rate × 3 × Triple Overtime Hours
Total Gross = Regular + Overtime + Bonuses + Commissions + Other
```

### Tax Calculations (ISR)

The system uses official SAT ISR tables for 2025:

1. Determine tax bracket based on taxable income
2. Calculate base tax
3. Apply marginal rate to excess
4. Subtract employment subsidy (if applicable)

### IMSS Calculations

Employee contributions calculated on SBC (Salario Base de Cotizacion):

| Concept | Rate |
|---------|------|
| Enfermedad y Maternidad | 0.625% |
| Invalidez y Vida | 0.625% |
| Cesantia y Vejez | 1.125% |

Employer contributions include additional rates plus work risk class.

---

## Mexican Payroll Compliance

### Regulatory Compliance

- **LFT** (Ley Federal del Trabajo) - Labor law compliance
- **LIMSS** (Ley del IMSS) - Social security contributions
- **LISR** (Ley del ISR) - Income tax withholding
- **INFONAVIT** - Housing fund contributions
- **SAT CFDI 4.0** - Electronic invoice requirements
- **Nomina 1.2** - Payroll complement for CFDI

### Key Mexican Payroll Concepts

| Concept | Description |
|---------|-------------|
| UMA | Unit of Measurement and Update (for benefits calculation) |
| SMG | General Minimum Wage |
| SBC | Base Contribution Salary (for IMSS) |
| SDI | Integrated Daily Salary |
| Aguinaldo | Christmas bonus (minimum 15 days) |
| Prima Vacacional | Vacation premium (25% of vacation pay) |
| PTU | Profit sharing |
| ISR | Income tax (Impuesto Sobre la Renta) |
| IMSS | Mexican Social Security Institute |
| INFONAVIT | Housing fund |
| RFC | Federal Taxpayer Registry |
| CURP | Unique Population Registry Code |
| NSS | Social Security Number |

### Payment Frequencies Supported

| Frequency | Days | Periods/Year |
|-----------|------|--------------|
| Weekly | 7 | 52 |
| Biweekly | 15 | 24 |
| Monthly | 30/31 | 12 |

---

## Employee Types (Collar Types)

The system supports three employee classifications that affect payroll calculation:

### White Collar (Empleados de Confianza)

- Salaried employees (exempt from overtime)
- Administrative, managerial, professional roles
- Monthly or biweekly payment
- Full benefits package
- No overtime calculations

### Blue Collar (Obreros)

- Hourly workers
- Manual labor, production, operations
- Weekly payment typical
- Overtime eligible (double/triple time)
- IMSS mandatory

### Gray Collar (Mixto)

- Hybrid classification
- Technical, skilled trades, supervisors
- Biweekly payment typical
- Limited overtime eligibility
- Mixed benefit structure

---

## Reports and Exports

### Available Reports

1. **Payroll Summary Report**
   - Total income, deductions, net pay by period
   - Employee count and totals

2. **Concept Totals Report**
   - Breakdown by payroll concept
   - Income vs deduction totals

3. **Employee Payroll History**
   - Individual employee payroll records
   - Period-by-period history

### Export Formats

| Format | Description |
|--------|-------------|
| PDF | Payslip (Recibo de Nomina) |
| XML | CFDI 4.0 with Nomina 1.2 complement |
| JSON | API response format |

---

## Frontend Pages

### Authentication
- `/auth/login` - User login page

### Dashboard
- `/dashboard` - Main dashboard with payroll overview

### Employees
- `/employees` - Employee list
- `/employees/new` - Create new employee
- `/employees/:id` - View/edit employee

### Payroll
- `/payroll` - Payroll periods list
- `/payroll/process` - Process payroll
- `/payroll/:periodId` - Period details

### Configuration
- `/configuration` - System settings

---

## Development Guide

### Code Style

#### Backend (Go)
- Follow Go conventions and `gofmt`
- Use meaningful variable names
- Document exported functions

#### Frontend (TypeScript)
- Follow TypeScript strict mode
- Use React functional components
- Implement proper error handling

### Adding New Features

1. **Backend**
   - Add model in `internal/models/`
   - Add repository in `internal/repositories/`
   - Add service in `internal/services/`
   - Add handler in `internal/api/`
   - Register routes in `router.go`

2. **Frontend**
   - Add page in `app/`
   - Add components in `components/`
   - Update API client in `lib/`

### Running Tests

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
pnpm test
```

### Building for Production

```bash
# Backend
cd backend
go build -ldflags="-s -w" -o iris-payroll ./cmd/server

# Frontend
cd frontend
pnpm build
```

---

## Troubleshooting

### Common Issues

#### Backend won't start
- Check if port 8080 is available: `lsof -i :8080`
- Verify configuration files exist in `configs/`
- Check database file permissions

#### Frontend API errors
- Verify backend is running
- Check CORS settings in backend `.env`
- Verify `NEXT_PUBLIC_API_URL` in frontend `.env.local`

#### Database migrations fail
- Delete `iris_payroll.db` and restart (development only)
- Check GORM model definitions for errors

#### Authentication issues
- Verify JWT_SECRET is set
- Check token expiration settings
- Clear browser cookies and try again

### Logs

- Backend logs to stdout
- Frontend logs in browser console
- Enable `DEBUG=true` for verbose logging

---

## License

Proprietary - All rights reserved

---

## Support

For support, please contact the development team or open an issue in the project repository.

---

## Changelog

### Version 1.0.0 (2025-01)
- Initial release
- Employee management
- Payroll calculations (ISR, IMSS, INFONAVIT)
- PDF payslip generation
- CFDI 4.0 XML generation
- Multi-collar type support
- San Luis Potosi regional configuration
# IRIS
# IRIS

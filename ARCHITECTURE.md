# IRIS Payroll System - Architecture Documentation

## System Overview

The IRIS Payroll System is a full-stack application designed for comprehensive payroll and employee management, featuring two separate frontends and a unified backend API.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        Nginx Reverse Proxy                   │
│                                                              │
│  Port 80 (Admin)          │         Port 8081 (Employee)    │
└───────────┬───────────────┴──────────────┬─────────────────┘
            │                               │
    ┌───────▼────────┐              ┌──────▼──────────┐
    │   Frontend     │              │ Employee Portal │
    │  (Admin UI)    │              │  (Self-Service) │
    │   Next.js      │              │    Next.js      │
    │   Port 3000    │              │   Port 3000     │
    └───────┬────────┘              └──────┬──────────┘
            │                               │
            └───────────────┬───────────────┘
                            │
                    ┌───────▼────────┐
                    │   Backend API  │
                    │   Go + Gin     │
                    │   Port 8080    │
                    └───────┬────────┘
                            │
                    ┌───────▼────────┐
                    │    SQLite DB   │
                    │  /data/*.db    │
                    └────────────────┘
```

---

## Services Configuration

### 1. Backend API
- **Technology**: Go 1.24 + Gin Framework + GORM
- **Port**: 8080
- **Database**: SQLite
- **Container**: `iris-payroll-backend`
- **Health Check**: `/health` endpoint
- **Features**:
  - JWT Authentication (access + refresh tokens)
  - Role-based authorization (admin, hr, manager, supervisor, employee)
  - Multi-stage absence request approval workflow
  - Payroll calculation engine
  - File upload handling (uploads, PDFs)
  - RESTful API design

### 2. Frontend (Admin Portal)
- **Technology**: Next.js 13+ (App Router) + TypeScript
- **Port**: 3000 (internal), 80 (via nginx)
- **Container**: `iris-payroll-frontend`
- **URL**: `http://localhost`
- **Features**:
  - Employee management (CRUD)
  - Payroll period management
  - Payroll calculation and processing
  - Reports and analytics
  - User management (admin only)
  - Incidence tracking
  - Configuration management

### 3. Employee Portal (Self-Service)
- **Technology**: Next.js 13+ (App Router) + TypeScript
- **Port**: 3000 (internal), 8081 (via nginx)
- **Container**: `iris-employee-portal`
- **URL**: `http://localhost:8081`
- **Features**:
  - Personal dashboard with vacation balance
  - Absence request submission (vacation, sick leave, permissions)
  - Request tracking and history
  - Multi-stage approval workflow visualization
  - Calendar view of absences
  - Reports and analytics
  - Announcements system
  - Password change
  - Notifications center

### 4. Nginx Reverse Proxy
- **Image**: nginx:alpine
- **Ports**: 80 (admin), 8081 (employee)
- **Container**: `iris-payroll-nginx`
- **Configuration**: `/nginx/nginx.conf`
- **Features**:
  - Reverse proxy routing
  - Gzip compression
  - Security headers
  - Large file upload support (50MB)
  - Health check endpoints

---

## Database Schema

### Core Tables
- **companies** - Company information
- **users** - User accounts with roles
- **employees** - Employee records
- **payroll_periods** - Payroll calculation periods
- **departments** - Organizational departments
- **positions** - Job positions
- **cost_centers** - Cost allocation centers

### Payroll Tables
- **prenomina_metrics** - Pre-payroll calculations
- **payroll_details** - Individual payroll entries
- **payroll_concepts** - Perception/deduction types
- **payroll_concept_values** - Calculated values

### Absence Management
- **absence_requests** - Employee absence requests
- **incidences** - HR-tracked incidences
- **incidence_types** - Catalog of incidence types

### Supporting Tables
- **notifications** - User notifications
- **announcements** - Company announcements
- **evidence_files** - File uploads (medical notes, etc.)

---

## API Endpoints

### Authentication (`/api/v1/auth`)
- `POST /login` - User login
- `POST /register` - Company registration
- `POST /refresh` - Refresh access token
- `POST /logout` - User logout
- `POST /change-password` - Change password

### Employees (`/api/v1/employees`)
- `GET /` - List employees
- `POST /` - Create employee
- `GET /:id` - Get employee details
- `PUT /:id` - Update employee
- `DELETE /:id` - Delete employee
- `GET /:id/vacation-balance` - Get vacation balance
- `GET /:id/absence-summary` - Get absence summary
- `GET /:id/portal-user` - Get portal user account
- `POST /:id/portal-user` - Create portal user account
- `PUT /:id/portal-user` - Update portal user account
- `DELETE /:id/portal-user` - Delete portal user account

### Absence Requests (`/api/v1/absence-requests`)
- `GET /` - List absence requests
- `POST /` - Create absence request
- `GET /my-requests` - Get current user's requests
- `GET /:id` - Get request details
- `POST /:id/approve/supervisor` - Supervisor approval
- `POST /:id/approve/manager` - Manager approval
- `POST /:id/approve/hr` - HR approval
- `POST /:id/decline` - Decline request
- `GET /counts` - Get pending counts by role
- `GET /pending/supervisor` - Get supervisor pending requests
- `GET /pending/manager` - Get manager pending requests
- `GET /pending/hr` - Get HR pending requests

### Announcements (`/api/v1/announcements`)
- `GET /` - List active announcements
- `GET /:id` - Get announcement details
- `POST /` - Create announcement (roles: admin, hr, manager, supervisor)
- `DELETE /:id` - Delete announcement
- `POST /:id/read` - Mark as read
- `GET /unread-count` - Get unread count

### Notifications (`/api/v1/notifications`)
- `GET /` - Get user notifications
- `GET /unread-count` - Get unread count
- `POST /:id/read` - Mark as read
- `POST /read-all` - Mark all as read

### Payroll (`/api/v1/payroll`)
- `GET /periods` - List payroll periods
- `POST /periods` - Create payroll period
- `POST /calculate` - Calculate payroll
- `GET /summary/:period_id` - Get payroll summary

---

## User Roles & Permissions

### Admin
- Full system access
- User management
- Company configuration
- All payroll and HR functions

### HR (Human Resources)
- Employee management
- Incidence tracking
- Absence request approval (final stage)
- Create announcements
- View all reports

### Manager
- Absence request approval (2nd stage)
- Team oversight
- View team reports
- Create announcements

### Supervisor
- Absence request approval (1st stage)
- Direct report management
- Create announcements

### Employee
- Submit absence requests
- View own data
- View announcements
- Change password

### Combined Roles
- `hr_and_pr` - HR + Payroll responsibilities
- `sup_and_gm` - Supervisor + General Manager (skips approval stages)

---

## Absence Request Workflow

```
┌────────────────┐
│    Employee    │
│  Submits       │
│  Request       │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│   Supervisor   │──── Approve ───┐
│   Review       │                │
└────────┬───────┘                │
         │ Decline                │
         ▼                        ▼
    [DECLINED]          ┌────────────────┐
                        │    Manager     │──── Approve ───┐
                        │    Review      │                │
                        └────────┬───────┘                │
                                 │ Decline                │
                                 ▼                        ▼
                            [DECLINED]          ┌────────────────┐
                                                │       HR       │
                                                │    Review      │
                                                └────────┬───────┘
                                                         │
                                    ┌────────────────────┼────────────────────┐
                                    │                    │                    │
                                    ▼                    ▼                    ▼
                              [APPROVED]           [DECLINED]           [REQUIRES_INFO]
```

---

## Environment Variables

### Backend
- `GIN_MODE` - Gin framework mode (debug/release)
- `JWT_SECRET` - JWT signing secret
- `JWT_REFRESH_SECRET` - Refresh token secret
- `DATABASE_URL` - SQLite database path
- `UPLOAD_PATH` - File upload directory
- `PDF_OUTPUT_PATH` - PDF generation directory
- `ALLOWED_ORIGINS` - CORS allowed origins

### Frontend/Employee Portal
- `NEXT_PUBLIC_API_URL` - Backend API URL
- `NODE_ENV` - Node environment (development/production)

---

## Docker Volumes

### iris-data
- **Type**: Local volume
- **Purpose**: Persistent data storage
- **Contents**:
  - SQLite database file
  - Uploaded files (evidence, documents)
  - Generated PDFs (payslips, reports)

---

## Security Features

1. **Authentication**
   - JWT-based authentication
   - Access + refresh token strategy
   - Secure password hashing (bcrypt)

2. **Authorization**
   - Role-based access control (RBAC)
   - Endpoint-level permissions
   - Multi-stage approval workflow

3. **HTTP Security**
   - CORS configuration
   - Security headers (X-Frame-Options, X-Content-Type-Options, X-XSS-Protection)
   - Request size limits
   - Input validation

4. **Data Protection**
   - Password complexity requirements
   - Soft deletes for data retention
   - Audit trails (created_by, updated_by, timestamps)

---

## Development Workflow

### Local Development (Without Docker)
```bash
# Backend
cd backend
go run cmd/api/main.go

# Frontend (Admin)
cd frontend
npm run dev

# Employee Portal
cd employee-portal
npm run dev
```

### Docker Development
```bash
# Start all services
bash scripts/docker-start.sh

# Stop all services
bash scripts/docker-stop.sh

# View logs
bash scripts/docker-logs.sh

# Restart services
bash scripts/docker-restart.sh
```

### Building for Production
```bash
# Build all services
docker compose build

# Start in production mode
docker compose up -d
```

---

## Monitoring & Health Checks

### Backend Health Check
- **Endpoint**: `http://localhost:8080/health`
- **Response**: `{"status": "ok", "service": "iris-payroll-backend"}`
- **Interval**: Every 30 seconds
- **Timeout**: 10 seconds

### Nginx Health Check
- **Admin Portal**: `http://localhost/health`
- **Employee Portal**: `http://localhost:8081/health`
- **Response**: `healthy`

---

## File Structure

```
iris-payroll-system/
├── backend/                 # Go backend API
│   ├── cmd/api/            # Main application
│   ├── internal/           # Internal packages
│   │   ├── api/           # HTTP handlers
│   │   ├── models/        # Data models
│   │   ├── services/      # Business logic
│   │   ├── repositories/  # Data access
│   │   ├── middleware/    # HTTP middleware
│   │   └── database/      # DB migrations
│   └── Dockerfile
│
├── frontend/               # Admin portal (Next.js)
│   ├── app/               # App router pages
│   ├── components/        # React components
│   ├── lib/              # Utilities & API client
│   └── Dockerfile
│
├── employee-portal/        # Employee self-service (Next.js)
│   ├── app/               # App router pages
│   ├── components/        # React components
│   ├── lib/              # Utilities & API client
│   └── Dockerfile
│
├── nginx/                  # Nginx configuration
│   └── nginx.conf
│
├── scripts/                # Utility scripts
│   ├── docker-start.sh
│   ├── docker-stop.sh
│   ├── docker-restart.sh
│   └── docker-logs.sh
│
├── docker-compose.yml      # Docker Compose configuration
└── ARCHITECTURE.md         # This file
```

---

## Recent Enhancements

### Employee Portal Features (Latest)
1. ✅ **Announcements System**
   - Company-wide communication
   - Image upload support
   - Scope filtering (ALL/DEPARTMENT/TEAM)
   - Expiration tracking

2. ✅ **Vacation Balance Widget**
   - Real-time balance calculation
   - Mexican labor law compliance
   - Visual progress indicators

3. ✅ **Password Change**
   - Secure dialog interface
   - Validation and confirmation
   - Error handling

4. ✅ **Schedule Calendar**
   - Monthly calendar view
   - Color-coded absences
   - Month navigation
   - Quick statistics

5. ✅ **Reports & Analytics Dashboard**
   - Key metrics (total, pending, approved, declined)
   - Monthly trend charts
   - Request type distribution
   - Processing time analytics
   - Approval rate tracking

---

## Performance Considerations

1. **Database Optimization**
   - Indexed columns (employee_number, rfc, curp, email)
   - Soft deletes for data retention
   - Query optimization via GORM

2. **Caching Strategy**
   - Static asset caching (nginx)
   - Gzip compression
   - Next.js build optimization

3. **Scalability**
   - Stateless backend (horizontally scalable)
   - SQLite for single-instance deployment
   - Can migrate to PostgreSQL/MySQL for high-load scenarios

---

## Support & Maintenance

### Logs Location
- **Backend**: Docker logs via `docker logs iris-payroll-backend`
- **Frontend**: Docker logs via `docker logs iris-payroll-frontend`
- **Employee Portal**: Docker logs via `docker logs iris-employee-portal`
- **Nginx**: `/var/log/nginx/access.log` and `/var/log/nginx/error.log`

### Backup Recommendations
1. Regular backups of `/data` volume
2. Database exports before major updates
3. Configuration file versioning

### Common Operations
```bash
# View all container logs
docker compose logs -f

# Restart specific service
docker compose restart backend

# Access backend container shell
docker exec -it iris-payroll-backend sh

# Database backup
docker cp iris-payroll-backend:/data/iris_payroll.db ./backup/

# View network configuration
docker network inspect iris-network
```

---

## Version Information

- **Backend**: Go 1.24 + Gin + GORM
- **Frontend**: Next.js 13+ + TypeScript + Tailwind CSS
- **Database**: SQLite 3
- **Proxy**: Nginx Alpine
- **Node**: v20 Alpine

---

*Last Updated: December 2025*
*System Version: 1.0.0*

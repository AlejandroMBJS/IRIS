# Quick Start Guide - IRIS Talent Payroll System

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.18+**: Download from https://golang.org/dl/
- **Node.js 18+**: Download from https://nodejs.org/
- **pnpm**: Install with `npm install -g pnpm`

## Step 1: Backend Setup

1. Open a terminal and navigate to the backend directory:
```bash
cd backend
```

2. Install Go dependencies:
```bash
go mod tidy
```

3. (Optional) Copy the example environment file:
```bash
cp .env.example .env
```

4. Start the backend server:
```bash
go run cmd/server/main.go
```

The backend should now be running on `http://localhost:8080`

## Step 2: Frontend Setup

1. Open a **new terminal** and navigate to the frontend directory:
```bash
cd frontend
```

2. Install Node.js dependencies:
```bash
pnpm install
```

3. (Optional) Create environment file:
```bash
cp .env.local.example .env.local
```

4. Start the development server:
```bash
pnpm dev
```

The frontend should now be running on `http://localhost:3000`

## Step 3: Access the Application

1. Open your web browser and go to: `http://localhost:3000`
2. You should see the login page
3. Click "Contacta al administrador" to register a new company

## Default Configuration

- **Database**: SQLite (file-based, no setup required)
- **Location**: San Luis Potosí, Mexico
- **Fiscal Year**: 2025
- **Currency**: MXN (Mexican Peso)

## Testing the Payroll Configuration

To verify the payroll configuration is loaded correctly:

```bash
cd backend
go run test_config.go
```

This will display:
- Fiscal Year
- UMA Daily Value
- Default Minimum Wage
- State Payroll Tax Rate

## Production Build

### Backend
```bash
cd backend
go build -o iris-payroll-backend cmd/server/main.go
./iris-payroll-backend
```

### Frontend
```bash
cd frontend
pnpm build
pnpm start
```

## Troubleshooting

### Backend Issues

**Problem**: `go: command not found`
- **Solution**: Install Go from https://golang.org/dl/

**Problem**: Port 8080 already in use
- **Solution**: Change `SERVER_PORT` in `.env` file or kill the process using port 8080

### Frontend Issues

**Problem**: `pnpm: command not found`
- **Solution**: Install pnpm with `npm install -g pnpm`

**Problem**: Port 3000 already in use
- **Solution**: The dev server will automatically try port 3001, or you can specify a port:
```bash
pnpm dev -p 3001
```

**Problem**: Build errors
- **Solution**: Delete `node_modules` and `.next` folders, then run:
```bash
pnpm install
pnpm dev
```

## Next Steps

1. **Configure Payroll**: Review and update JSON files in `backend/configs/`
2. **Add Employees**: Use the employee management interface
3. **Process Payroll**: Calculate payroll for your employees
4. **Generate Reports**: View compliance and analytics reports

## Support

For detailed documentation, see `README.md`

For issues or questions, contact your system administrator.

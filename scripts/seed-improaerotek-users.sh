#!/bin/bash

# Seed script for Improaerotek default users
# Creates admin and role-specific users

API_URL="http://localhost:80/api/v1"
PASSWORD="Password123@"
DOMAIN="@improaerotek.com"

echo "=========================================="
echo "IRIS - Improaerotek Users Seed"
echo "=========================================="

# Function to extract token from JSON response (no jq needed)
extract_token() {
  echo "$1" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4
}

# Function to extract email or error from response
extract_result() {
  local response="$1"
  if echo "$response" | grep -q '"email"'; then
    echo "$response" | grep -o '"email":"[^"]*"' | head -1 | cut -d'"' -f4
  elif echo "$response" | grep -q '"error"'; then
    echo "ERROR: $(echo "$response" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)"
  else
    echo "$response"
  fi
}

# Step 1: Register Admin User (creates company too)
echo ""
echo "=== Step 1: Registering Admin User ==="
ADMIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Improaerotek SA de CV",
    "company_rfc": "IAT123456789",
    "email": "admin'"$DOMAIN"'",
    "password": "'"$PASSWORD"'",
    "full_name": "Administrador Sistema",
    "role": "admin"
  }')

extract_result "$ADMIN_RESPONSE"
TOKEN=$(extract_token "$ADMIN_RESPONSE")

if [ -z "$TOKEN" ]; then
  echo "Admin might already exist. Trying login..."
  ADMIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email": "admin'"$DOMAIN"'", "password": "'"$PASSWORD"'"}')
  TOKEN=$(extract_token "$ADMIN_RESPONSE")
fi

if [ -z "$TOKEN" ]; then
  echo "ERROR: Could not get admin token"
  echo "Response: $ADMIN_RESPONSE"
  exit 1
fi

echo "Admin Token obtained: ${TOKEN:0:50}..."

# Step 2: Create HR White Collar Users
echo ""
echo "=== Step 2: Creating HR White Collar Users ==="

echo -n "Creating betsabe_cortes... "
RESPONSE=$(curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "betsabe_cortes'"$DOMAIN"'",
    "password": "'"$PASSWORD"'",
    "full_name": "Betsabe Cortes",
    "role": "hr_white"
  }')
extract_result "$RESPONSE"

echo -n "Creating ana_solis... "
RESPONSE=$(curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "ana_solis'"$DOMAIN"'",
    "password": "'"$PASSWORD"'",
    "full_name": "Ana Solis",
    "role": "hr_white"
  }')
extract_result "$RESPONSE"

# Step 3: Create HR Blue & Gray Collar Users
echo ""
echo "=== Step 3: Creating HR Blue & Gray Collar Users ==="

echo -n "Creating jennifer_barcenas... "
RESPONSE=$(curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jennifer_barcenas'"$DOMAIN"'",
    "password": "'"$PASSWORD"'",
    "full_name": "Jennifer Barcenas",
    "role": "hr_blue_gray"
  }')
extract_result "$RESPONSE"

echo -n "Creating agustina_lopez... "
RESPONSE=$(curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "agustina_lopez'"$DOMAIN"'",
    "password": "'"$PASSWORD"'",
    "full_name": "Agustina Lopez",
    "role": "hr_blue_gray"
  }')
extract_result "$RESPONSE"

# Step 4: Create Payroll Users
echo ""
echo "=== Step 4: Creating Payroll Users ==="

echo -n "Creating gabriela_alvarado... "
RESPONSE=$(curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "gabriela_alvarado'"$DOMAIN"'",
    "password": "'"$PASSWORD"'",
    "full_name": "Gabriela Alvarado",
    "role": "payroll_staff"
  }')
extract_result "$RESPONSE"

echo -n "Creating nancy_gonzalez... "
RESPONSE=$(curl -s -X POST "$API_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nancy_gonzalez'"$DOMAIN"'",
    "password": "'"$PASSWORD"'",
    "full_name": "Nancy Gonzalez",
    "role": "payroll_staff"
  }')
extract_result "$RESPONSE"

echo ""
echo "=========================================="
echo "Seed Complete!"
echo "=========================================="
echo ""
echo "Users created:"
echo "  ADMIN:"
echo "    - admin$DOMAIN"
echo ""
echo "  HR WHITE COLLAR:"
echo "    - betsabe_cortes$DOMAIN"
echo "    - ana_solis$DOMAIN"
echo ""
echo "  HR BLUE & GRAY:"
echo "    - jennifer_barcenas$DOMAIN"
echo "    - agustina_lopez$DOMAIN"
echo ""
echo "  PAYROLL:"
echo "    - gabriela_alvarado$DOMAIN"
echo "    - nancy_gonzalez$DOMAIN"
echo ""
echo "All passwords: $PASSWORD"
echo ""

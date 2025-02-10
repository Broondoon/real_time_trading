#!/bin/bash

# API Base URL
BASE_URL="http://localhost:8000"

# Sample User Credentials
USERNAME="testuser"
PASSWORD="securepassword"

echo "========================="
echo "Registering User"
echo "========================="
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/register" -H "Content-Type: application/json" \
    -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}")

echo "Response: $REGISTER_RESPONSE"
echo ""

echo "========================="
echo "Logging In"
echo "========================="
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" -H "Content-Type: application/json" \
    -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}")

echo "Response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | jq -r .token) # Extract JWT token using jq

if [[ "$TOKEN" == "null" ]]; then
    echo "Login failed. Exiting..."
    exit 1
fi

echo "Token Acquired: $TOKEN"
echo ""

echo "Curl command: curl -X GET \"$BASE_URL/protected/profile\" -H \"Authorization: Bearer $TOKEN\" -H \"Content-Type: application/json\""

echo "========================="
echo "Accessing Protected Route with Token"
echo "========================="
PROTECTED_RESPONSE=$(curl -s -X GET "$BASE_URL/protected/profile" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

echo "Response: $PROTECTED_RESPONSE"
echo ""

echo "========================="
echo "Attempting Protected Route Without Token (Should Fail)"
echo "========================="
NO_AUTH_RESPONSE=$(curl -s -X GET "$BASE_URL/protected/profile")

echo "Response: $NO_AUTH_RESPONSE"
echo ""

echo "Test Completed!"

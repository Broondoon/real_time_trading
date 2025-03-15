#!/bin/bash
# jmeter_to_curl.sh
# This script converts a JMeter test plan into a series of curl requests.
# It requires curl and jq to be installed.

# TestPlan variables from the JMX file:
BASE_URL="localhost"
BASE_PORT="3001"
register="/authentication/register"
login="/authentication/login"
getStockPrices="/transaction/getStockPrices"
getStockPortfolio="/transaction/getStockPortfolio"
placeStockOrder="/engine/placeStockOrder"
getStockTransactions="/transaction/getStockTransactions"
addMoneyToWallet="/transaction/addMoneyToWallet"
getWalletBalance="/transaction/getWalletBalance"
getWalletTransactions="/transaction/getWalletTransactions"
cancelStockTransaction="/engine/cancelStockTransaction"
createStock="/setup/createStock"
addStockToUser="/setup/addStockToUser"
invalidToken="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"

# Helper: delay of 500ms (simulating ConstantTimer)
delay() {
    sleep 0.5
}

##########################
# Request 1: Valid Register
##########################
echo "Request 1: Register (valid)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$register" \
  -H "Content-Type: application/json" \
  --data '{ "user_name": "VanguardETF", "password": "Vang@123", "name": "Vanguard Corp." }')
echo "Response: $response"

##########################
# Request 2: Failed Register
##########################
delay
echo "Request 2: Register (failed)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$register" \
  -H "Content-Type: application/json" \
  --data '{ "user_name": "VanguardETF", "password": "Comp@124", "name": "Vanguard Ltd." }')
echo "Response: $response"

##########################
# Request 3: Failed Login
##########################
delay
echo "Request 3: Login (failed)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$login" \
  -H "Content-Type: application/json" \
  --data '{ "user_name": "VanguardETF", "password": "Vang@1234" }')
echo "Response: $response"

##########################
# Request 4: Successful Login (extract compToken)
##########################
delay
echo "Request 4: Login (successful)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$login" \
  -H "Content-Type: application/json" \
  --data '{ "user_name": "VanguardETF", "password": "Vang@123" }')
echo "Response: $response"
compToken=$(echo "$response" | jq -r '.data.token')
echo "compToken: $compToken"

##########################
# Request 5: Create Google Stock (extract googleStockId)
##########################
delay
echo "Request 5: Create Google Stock"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$createStock" \
  -H "Content-Type: application/json" \
  -H "token: $compToken" \
  --data '{ "stock_name": "Google" }')
echo "Response: $response"
googleStockId=$(echo "$response" | jq -r '.data.stock_id')
echo "googleStockId: $googleStockId"

##########################
# Request 6: Add Google Stock to User
##########################
delay
echo "Request 6: Add Google Stock to User"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$addStockToUser" \
  -H "Content-Type: application/json" \
  -H "token: $compToken" \
  --data "{ \"stock_id\": \"$googleStockId\", \"quantity\": 550 }")
echo "Response: $response"

##########################
# Request 7: Create Apple Stock (extract appleStockId)
##########################
delay
echo "Request 7: Create Apple Stock"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$createStock" \
  -H "Content-Type: application/json" \
  -H "token: $compToken" \
  --data '{ "stock_name": "Apple" }')
echo "Response: $response"
appleStockId=$(echo "$response" | jq -r '.data.stock_id')
echo "appleStockId: $appleStockId"

##########################
# Request 8: Add Apple Stock to User
##########################
delay
echo "Request 8: Add Apple Stock to User"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$addStockToUser" \
  -H "Content-Type: application/json" \
  -H "token: $compToken" \
  --data "{ \"stock_id\": \"$appleStockId\", \"quantity\": 350 }")
echo "Response: $response"

##########################
# Request 9: Get Stock Portfolio (for compToken)
##########################
delay
echo "Request 9: Get Stock Portfolio"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getStockPortfolio" \
  -H "Content-Type: application/json" \
  -H "token: $compToken")
echo "Response: $response"
# (Expected: Contains Google and Apple stock details)

##########################
# Request 10: Place Stock Order (Google, LIMIT sell)
##########################
echo "Request 10: Place Stock Order for Google (LIMIT sell)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$placeStockOrder" \
  -H "Content-Type: application/json" \
  -H "token: $compToken" \
  --data "{ \"stock_id\": \"$googleStockId\", \"is_buy\": false, \"order_type\": \"LIMIT\", \"quantity\": 550, \"price\": 135 }")
echo "Response: $response"

##########################
# Request 11: Place Stock Order (Apple, LIMIT sell)
##########################
echo "Request 11: Place Stock Order for Apple (LIMIT sell)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$placeStockOrder" \
  -H "Content-Type: application/json" \
  -H "token: $compToken" \
  --data "{ \"stock_id\": \"$appleStockId\", \"is_buy\": false, \"order_type\": \"LIMIT\", \"quantity\": 350, \"price\": 140 }")
echo "Response: $response"

##########################
# Request 12: Get Stock Portfolio (expect empty)
##########################
delay
echo "Request 12: Get Stock Portfolio (expecting empty list)"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getStockPortfolio" \
  -H "Content-Type: application/json" \
  -H "token: $compToken")
echo "Response: $response"

##########################
# Request 13: Get Stock Transactions (extract transaction IDs)
##########################
echo "Request 13: Get Stock Transactions"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getStockTransactions" \
  -H "Content-Type: application/json" \
  -H "token: $compToken")
echo "Response: $response"
# Extract transaction IDs (for later cancellation and verification)
googleCompStockTxId=$(echo "$response" | jq -r '.data[0].stock_tx_id')
# For simplicity, we set googleUserStockTxId the same as googleCompStockTxId
googleUserStockTxId="$googleCompStockTxId"
googleUserWalletTxId=$(echo "$response" | jq -r '.data[0].wallet_tx_id')
echo "googleCompStockTxId: $googleCompStockTxId"
echo "googleUserStockTxId: $googleUserStockTxId"
echo "googleUserWalletTxId: $googleUserWalletTxId"

##########################
# Request 14: Register FinanceGuru
##########################
echo "Request 14: Register FinanceGuru"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$register" \
  -H "Content-Type: application/json" \
  --data '{ "user_name": "FinanceGuru", "password": "Fguru@2024", "name": "The Finance Guru" }')
echo "Response: $response"

##########################
# Request 15: Login FinanceGuru (extract user1Token)
##########################
delay
echo "Request 15: Login FinanceGuru"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$login" \
  -H "Content-Type: application/json" \
  --data '{ "user_name": "FinanceGuru", "password": "Fguru@2024" }')
echo "Response: $response"
user1Token=$(echo "$response" | jq -r '.data.token')
echo "user1Token: $user1Token"

##########################
# Request 16: Get Stock Prices (FinanceGuru)
##########################
echo "Request 16: Get Stock Prices"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getStockPrices" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token")
echo "Response: $response"

##########################
# Request 17: Add Money to Wallet (FinanceGuru)
##########################
echo "Request 17: Add Money to Wallet"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$addMoneyToWallet" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token" \
  --data '{ "amount": 10000 }')
echo "Response: $response"

##########################
# Request 18: Get Wallet Balance (FinanceGuru)
##########################
delay
echo "Request 18: Get Wallet Balance"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getWalletBalance" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token")
echo "Response: $response"

##########################
# Request 19: Place Stock Order (FinanceGuru, MARKET buy for Google)
##########################
echo "Request 19: Place Stock Order for Google (MARKET buy)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$placeStockOrder" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token" \
  --data "{ \"stock_id\": \"$googleStockId\", \"is_buy\": true, \"order_type\": \"MARKET\", \"quantity\": 10 }")
echo "Response: $response"

##########################
# Request 20: Get Stock Transactions (FinanceGuru)
##########################
delay
echo "Request 20: Get Stock Transactions (FinanceGuru)"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getStockTransactions" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token")
echo "Response: $response"
# (Optionally update googleUserStockTxId / googleUserWalletTxId here if needed)

##########################
# Request 21: Get Wallet Transactions (FinanceGuru)
##########################
echo "Request 21: Get Wallet Transactions"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getWalletTransactions" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token")
echo "Response: $response"

##########################
# Request 22: Get Wallet Balance (FinanceGuru, after transaction)
##########################
echo "Request 22: Get Wallet Balance"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getWalletBalance" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token")
echo "Response: $response"

##########################
# Request 23: Get Stock Portfolio (FinanceGuru)
##########################
echo "Request 23: Get Stock Portfolio (FinanceGuru)"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getStockPortfolio" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token")
echo "Response: $response"

##########################
# Request 24: Cancel Stock Order (FinanceGuru, using googleCompStockTxId)
##########################
echo "Request 24: Cancel Stock Order (FinanceGuru)"
response=$(curl -s -X POST "http://$BASE_URL:$BASE_PORT$cancelStockTransaction" \
  -H "Content-Type: application/json" \
  -H "token: $user1Token" \
  --data "{ \"stock_tx_id\": \"$googleCompStockTxId\" }")
echo "Response: $response"

##########################
# Request 25: Get Stock Transactions (using compToken)
##########################
echo "Request 25: Get Stock Transactions (for compToken)"
response=$(curl -s -X GET "http://$BASE_URL:$BASE_PORT$getStockTransactions" \
  -H "Content-Type: application/json" \
  -H "token: $compToken")
echo "Response: $response"

echo "Test plan complete."


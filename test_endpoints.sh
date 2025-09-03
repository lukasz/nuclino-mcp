#!/bin/bash

# Test script to verify Nuclino API endpoints
# Usage: NUCLINO_API_KEY=your_key ./test_endpoints.sh

if [ -z "$NUCLINO_API_KEY" ]; then
    echo "Error: NUCLINO_API_KEY environment variable is required"
    exit 1
fi

BASE_URL="https://api.nuclino.com"

echo "🔍 Testing Nuclino API endpoints"
echo "================================"

# Function to make API calls
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local description="$4"
    
    echo
    echo "Testing: $description"
    echo "URL: $method $BASE_URL$endpoint"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: $NUCLINO_API_KEY" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: $NUCLINO_API_KEY" \
            -H "Content-Type: application/json" \
            "$BASE_URL$endpoint")
    fi
    
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    echo "Status: $status_code"
    if [ ${#body} -gt 200 ]; then
        echo "Body (truncated): $(echo "$body" | head -c 200)..."
    else
        echo "Body: $body"
    fi
    
    if [ "$status_code" -ge 400 ]; then
        echo "❌ FAILED"
    else
        echo "✅ SUCCESS"
    fi
    echo "----------------------------------------"
}

# Test 1: List workspaces (this should work based on your report)
test_endpoint "GET" "/v0/workspaces" "" "List workspaces"

# Test 2: List teams
test_endpoint "GET" "/v0/teams" "" "List teams"

# Get first workspace ID from response if possible
WORKSPACE_ID=$(curl -s -H "Authorization: $NUCLINO_API_KEY" \
    "$BASE_URL/v0/workspaces" | \
    python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    if 'data' in data and 'results' in data['data'] and len(data['data']['results']) > 0:
        print(data['data']['results'][0]['id'])
    elif 'results' in data and len(data['results']) > 0:
        print(data['results'][0]['id'])
except:
    pass
" 2>/dev/null)

if [ -n "$WORKSPACE_ID" ]; then
    echo "🎯 Found workspace ID: $WORKSPACE_ID"
    echo "Testing workspace-specific endpoints..."
    
    # Test 3: Get specific workspace
    test_endpoint "GET" "/v0/workspaces/$WORKSPACE_ID" "" "Get workspace details"
    
    # Test 4: List items in workspace - różne warianty endpointu
    test_endpoint "GET" "/v0/workspaces/$WORKSPACE_ID/items" "" "List workspace items (variant 1)"
    test_endpoint "GET" "/v0/items?workspaceId=$WORKSPACE_ID" "" "List workspace items (variant 2)"
    
    # Test 5: List collections in workspace
    test_endpoint "GET" "/v0/workspaces/$WORKSPACE_ID/collections" "" "List workspace collections"
    
    # Test 6: Search items
    test_endpoint "POST" "/v0/items/search" '{"query":"test","workspaceId":"'$WORKSPACE_ID'"}' "Search items"
    
    # Test 7: Create item with workspaceId
    test_endpoint "POST" "/v0/items" '{"title":"Test Item API","content":"Test content","workspaceId":"'$WORKSPACE_ID'"}' "Create item with workspaceId"
    
    # Test 8: List files in workspace
    test_endpoint "GET" "/v0/workspaces/$WORKSPACE_ID/files" "" "List workspace files"
    
else
    echo "❌ Could not extract workspace ID from API response"
    echo "Skipping workspace-specific tests"
fi

# Test other endpoints that don't require workspace ID

# Test 9: Try to get current user (should fail based on your analysis)
test_endpoint "GET" "/v0/user" "" "Get current user (should fail)"

echo
echo "🎉 API endpoint testing completed!"
echo "Review the results above to see which endpoints work and which don't."
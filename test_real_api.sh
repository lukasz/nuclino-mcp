#!/bin/bash

# Test official Nuclino API endpoints with real API key
# Based on documentation: https://help.nuclino.com/fa38d15f-items-and-collections

API_KEY="8QiLxmz5OadkPXyO+ohXfpgKC6R9NDtYE6fbReMR"
BASE_URL="https://api.nuclino.com"

echo "🧪 Testing Official Nuclino API Endpoints"
echo "========================================"

# Function to make API calls with proper error handling
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local description="$4"
    
    echo
    echo "📍 Testing: $description"
    echo "   Method: $method"
    echo "   URL: $BASE_URL$endpoint"
    
    if [ -n "$data" ]; then
        echo "   Data: $data"
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: $API_KEY" \
            -H "Content-Type: application/json" \
            -H "Accept: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: $API_KEY" \
            -H "Content-Type: application/json" \
            -H "Accept: application/json" \
            "$BASE_URL$endpoint" 2>/dev/null)
    fi
    
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    echo "   Status: $status_code"
    
    # Pretty print JSON if possible
    if command -v python3 >/dev/null 2>&1; then
        if echo "$body" | python3 -m json.tool >/dev/null 2>&1; then
            if [ ${#body} -gt 500 ]; then
                echo "   Response (truncated):"
                echo "$body" | python3 -m json.tool | head -20
                echo "   ... (truncated)"
            else
                echo "   Response:"
                echo "$body" | python3 -m json.tool
            fi
        else
            echo "   Response: $body"
        fi
    else
        echo "   Response: $body"
    fi
    
    if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
        echo "   ✅ SUCCESS"
    elif [ "$status_code" -eq 404 ]; then
        echo "   ❌ NOT FOUND (endpoint may not exist)"
    elif [ "$status_code" -eq 401 ]; then
        echo "   ❌ UNAUTHORIZED (API key issue)"
    else
        echo "   ❌ FAILED"
    fi
    echo "   ----------------------------------------"
}

# Test 1: Get teams (should work - basic auth test)
test_endpoint "GET" "/v0/teams" "" "Get teams (auth test)"

# Test 2: Get workspaces - WRONG endpoint we've been using
test_endpoint "GET" "/v0/workspaces" "" "Get workspaces (our current endpoint - may be wrong)"

# Test 3: Get items without parameters (should show what's available)
test_endpoint "GET" "/v0/items" "" "Get items (no parameters)"

# Get first workspace/team ID for further tests
WORKSPACE_ID=$(curl -s -H "Authorization: $API_KEY" "$BASE_URL/v0/teams" | \
    python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    if 'data' in data and 'results' in data['data'] and len(data['data']['results']) > 0:
        # Try to get first workspace from first team
        team = data['data']['results'][0]
        if 'workspaces' in team and len(team['workspaces']) > 0:
            print(team['workspaces'][0]['id'])
        elif 'id' in team:
            print(team['id'])  # Use team ID as fallback
    elif 'results' in data and len(data['results']) > 0:
        item = data['results'][0]
        if 'id' in item:
            print(item['id'])
except:
    pass
" 2>/dev/null)

if [ -n "$WORKSPACE_ID" ]; then
    echo
    echo "🎯 Found workspace/team ID: $WORKSPACE_ID"
    echo "Testing workspace-specific endpoints..."
    
    # Test 4: Get items by workspaceId (official way)
    test_endpoint "GET" "/v0/items?workspaceId=$WORKSPACE_ID" "" "Get items by workspaceId (official)"
    
    # Test 5: Get items by teamId (alternative)
    test_endpoint "GET" "/v0/items?teamId=$WORKSPACE_ID" "" "Get items by teamId (alternative)"
    
    # Test 6: Search items (official way)
    test_endpoint "GET" "/v0/items?workspaceId=$WORKSPACE_ID&search=test" "" "Search items (official)"
    
    # Test 7: Create item with workspaceId (what should work)
    test_endpoint "POST" "/v0/items" '{"workspaceId":"'$WORKSPACE_ID'","title":"API Test Item","content":"Created via direct API test"}' "Create item with workspaceId (official)"
    
    # Get an existing item ID for update/delete tests
    ITEM_ID=$(curl -s -H "Authorization: $API_KEY" "$BASE_URL/v0/items?workspaceId=$WORKSPACE_ID&limit=1" | \
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
    
    if [ -n "$ITEM_ID" ]; then
        echo
        echo "🎯 Found item ID for testing: $ITEM_ID"
        
        # Test 8: Get single item
        test_endpoint "GET" "/v0/items/$ITEM_ID" "" "Get single item (official)"
        
        # Test 9: Update item (PUT not PATCH!)
        test_endpoint "PUT" "/v0/items/$ITEM_ID" '{"title":"Updated via API test"}' "Update item with PUT (official)"
        
        # Test 10: Delete item (for cleanup)
        # test_endpoint "DELETE" "/v0/items/$ITEM_ID" "" "Delete item (cleanup)"
        echo "   ⚠️  Skipping delete test to preserve data"
    else
        echo "   ❌ No items found for update/delete tests"
    fi
else
    echo "   ❌ Could not extract workspace/team ID for further tests"
fi

# Test other endpoints that might be different than expected

# Test 11: Try to find workspace endpoint
test_endpoint "GET" "/v0/workspaces?limit=5" "" "Get workspaces with parameters"

echo
echo "🎉 API testing completed!"
echo "Review the results above to see which endpoints work and which don't."
echo "This will help us fix the MCP server implementation."
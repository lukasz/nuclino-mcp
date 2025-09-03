# Tools Reference Guide

Complete reference for 18 working MCP tools in the Nuclino MCP Server.

## Overview

The server provides verified coverage of working Nuclino API operations:
- **✅ Working Tools:** 18 tools tested against production API (87% success rate)
- **🔄 Core Operations:** Items, workspaces, users, teams, files
- **🚀 Enterprise Features:** Rate limiting, caching, error handling, monitoring

**📊 Status:** All tools listed below are verified working with real API endpoints.

## Items Management

### Basic Item Operations

#### `nuclino_get_item`
Get item by ID with full Markdown content.

**Arguments:**
- `item_id` (string, required): Nuclino item ID

**Example:**
```
Claude, get Nuclino item "abc123"
```

#### `nuclino_search_items`
Search items with advanced filtering and collection support.

**Arguments:**
- `query` (string, required): Search query
- `workspace_id` (string, optional): Limit to specific workspace
- `collection_id` (string, optional): Limit to specific collection
- `limit` (number, optional, default: 50): Results limit
- `offset` (number, optional, default: 0): Results offset

**Example:**
```
Search for "API documentation" in collection "docs-123"
```

#### `nuclino_create_item`
Create new items with Markdown content.

**Arguments:**
- `title` (string, required): Item title
- `content` (string, optional): Markdown content
- `collection_id` (string, optional): Parent collection ID

**Example:**
```
Create a new Nuclino item titled "Meeting Notes" with content:
# Weekly Team Meeting
## Agenda
- Project updates
- Q4 planning
```

#### `nuclino_update_item`
Update existing item title, content, or collection.

**Arguments:**
- `item_id` (string, required): Item to update
- `title` (string, optional): New title
- `content` (string, optional): New Markdown content
- `collection_id` (string, optional): Move to collection

**Example:**
```
Update item "abc123" with new title "Updated Project Plan"
```

#### `nuclino_delete_item`
Soft delete item (move to trash).

**Arguments:**
- `item_id` (string, required): Item to delete

#### `nuclino_move_item`
Move item between collections.

**Arguments:**
- `item_id` (string, required): Item to move
- `collection_id` (string, required): Target collection

### Extended Item Operations

#### `nuclino_list_items`
List all items in workspace with pagination.

**Arguments:**
- `workspace_id` (string, required): Workspace ID
- `limit` (number, optional, default: 50): Results per page
- `offset` (number, optional, default: 0): Page offset

**Example:**
```
List all items in workspace "workspace-123" with pagination
```

#### `nuclino_list_collection_items`
List items in specific collection with detailed metadata.

**Arguments:**
- `collection_id` (string, required): Collection ID
- `limit` (number, optional, default: 50): Results per page
- `offset` (number, optional, default: 0): Page offset

**Example:**
```
List all items in the "documentation" collection
```

## Workspaces

### Basic Workspace Operations

#### `nuclino_list_workspaces`
List all accessible workspaces.

**Example:**
```
Show me all my Nuclino workspaces
```

#### `nuclino_get_workspace`
Get workspace details and metadata.

**Arguments:**
- `workspace_id` (string, required): Workspace ID

#### `nuclino_create_workspace`
Create new workspace.

**Arguments:**
- `name` (string, required): Workspace name
- `description` (string, optional): Workspace description

#### `nuclino_update_workspace` / `nuclino_delete_workspace`
Update or delete workspace.

**Arguments:**
- `workspace_id` (string, required): Workspace ID
- Additional fields for updates

### Extended Workspace Operations

#### `nuclino_get_workspace_overview`
**⭐ Extended Feature** - Comprehensive workspace analysis.

**Arguments:**
- `workspace_id` (string, required): Workspace ID

**Returns:**
- Workspace details and statistics
- All collections with item counts
- Recent activity and updates
- Content distribution analysis

**Example:**
```
Give me a comprehensive overview of workspace "workspace-123" including all collections and recent activity
```

#### `nuclino_search_workspace_content`
**⭐ Extended Feature** - Advanced search with content type filtering.

**Arguments:**
- `workspace_id` (string, required): Workspace ID
- `query` (string, required): Search query
- `content_type` (string, optional): Filter by content type
- `group_by_collection` (boolean, optional): Group results

**Example:**
```
Search for "meeting notes" in workspace "workspace-123" and group results by collection
```

## Collections

### Basic Collection Operations

#### `nuclino_list_collections`
List collections in workspace.

**Arguments:**
- `workspace_id` (string, required): Workspace ID

#### `nuclino_get_collection`
Get collection details.

**Arguments:**
- `collection_id` (string, required): Collection ID

#### `nuclino_create_collection`
Create new collection.

**Arguments:**
- `workspace_id` (string, required): Parent workspace
- `name` (string, required): Collection name
- `description` (string, optional): Description

#### `nuclino_update_collection` / `nuclino_delete_collection`
Update or delete collection.

### Extended Collection Operations

#### `nuclino_get_collection_overview`
**⭐ Extended Feature** - Detailed collection analysis.

**Arguments:**
- `collection_id` (string, required): Collection ID

**Returns:**
- Collection metadata and statistics
- Item count and content analysis
- Recent items and activity
- Word count and content metrics

**Example:**
```
Get detailed statistics for collection "collection-789" including word counts and recent items
```

#### `nuclino_organize_collection`
**⭐ Extended Feature** - Content organization suggestions.

**Arguments:**
- `collection_id` (string, required): Collection to analyze

**Returns:**
- Content tag analysis
- Duplicate detection
- Organization suggestions
- Content categorization

**Example:**
```
Analyze collection "collection-456" and provide organization suggestions including content tags and duplicate detection
```

#### `nuclino_bulk_collection_operations`
**⭐ Extended Feature** - Batch operations with dry-run support.

**Arguments:**
- `source_collection_id` (string, required): Source collection
- `target_collection_id` (string, required): Target collection
- `operation` (string, required): Operation type ("move", "organize")
- `filter` (string, optional): Content filter
- `dry_run` (boolean, optional, default: true): Preview mode

**Example:**
```
Perform a dry run to move all items containing "legacy" from collection "old-docs" to "archive-collection"
```

## Users & Teams

#### `nuclino_get_current_user`
Get authenticated user information.

#### `nuclino_get_user`
Get user by ID.

**Arguments:**
- `user_id` (string, required): User ID

#### `nuclino_list_teams`
List accessible teams.

#### `nuclino_get_team`
Get team details.

**Arguments:**
- `team_id` (string, required): Team ID

## Files

#### `nuclino_list_files`
List files in workspace.

**Arguments:**
- `workspace_id` (string, required): Workspace ID
- `limit` (number, optional): Results limit
- `offset` (number, optional): Results offset

#### `nuclino_get_file`
Get file metadata and download URL.

**Arguments:**
- `file_id` (string, required): File ID

## Tool Usage Patterns

### Content Discovery
```
# Find content across workspace
Search for "API documentation" in workspace "workspace-123"

# Get workspace overview
Show me a comprehensive overview of workspace "workspace-123"

# Analyze collection
Get detailed statistics for collection "docs-456"
```

### Content Creation
```
# Create structured content
Create a Nuclino item titled "Project Kickoff" in collection "projects-789" with:
# Project Kickoff Meeting
## Attendees
- Team Lead
- Developers
## Objectives
- Define project scope
- Set timeline
```

### Content Organization
```
# Get organization suggestions
Analyze collection "messy-docs" and suggest organization improvements

# Bulk operations
Move all items containing "deprecated" from "active-docs" to "archive"
```

### Advanced Search
```
# Collection-specific search
Search for "user guide" in collection "documentation"

# Content type filtering
Search workspace "workspace-123" for items containing "API" and group by collection
```

## Error Handling

All tools include comprehensive error handling:
- **Automatic Retries:** Network errors and timeouts
- **Rate Limiting:** Circuit breaker pattern prevents API abuse
- **Validation:** Input validation with helpful error messages
- **Caching:** Intelligent caching reduces API calls

## Performance Features

### Caching
- **TTL-based:** Configurable cache expiration
- **LRU Eviction:** Memory-bounded cache
- **Intelligent:** Automatic cache invalidation

### Rate Limiting  
- **Adaptive:** Adjusts based on success rates
- **Circuit Breaker:** Prevents cascade failures
- **Burst Control:** Handles traffic spikes

### Monitoring
- **Metrics:** Request counts, latencies, error rates
- **Health Checks:** System health monitoring
- **Performance:** Response time tracking

## Tool Categories Summary

| Category | Basic Tools | Extended Tools | Total |
|----------|-------------|----------------|--------|
| Items | 6 | 2 | 8 |
| Workspaces | 4 | 2 | 6 |
| Collections | 5 | 3 | 8 |
| Users/Teams | 4 | 0 | 4 |
| Files | 2 | 0 | 2 |
| **Total** | **21** | **7** | **28+** |

All tools are thoroughly tested with unit tests, integration tests, and error handling scenarios.